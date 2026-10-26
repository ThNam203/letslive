package live

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nareix/joy5/av"
	"github.com/nareix/joy5/format/flv"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/transcode/config"
)

// encodedReadyDeadline is how long the encoded renditions get, after the
// source became ready, to become ready themselves (ruling R6). A variable
// only so tests can shorten it.
var encodedReadyDeadline = 10 * time.Second

// encoderWiring is what the encoder side of a Session connects to.
type encoderWiring struct {
	stream    *Stream
	rec       *Recorder
	clk       *Clock
	src       SourceInfo
	ladder    []Rendition // index 0 is the source; the rest are encoded
	headers   []av.Packet // decoder configs, written to ffmpeg before any media
	hasAudio  bool
	transcode config.Transcode
}

// sessionEncoder feeds one ffmpeg process through a bounded queue and feeds
// its outputs into the encoded renditions. When ffmpeg fails in any way the
// encoded renditions are dropped from the stream and the source keeps going.
type sessionEncoder struct {
	ctx    context.Context
	id     string
	stream *Stream
	count  int // encoded renditions, ladder indexes 1..count
	enc    *Encoder

	queue      chan av.Packet
	halt       chan struct{} // closed once the queue must no longer be fed
	haltOnce   sync.Once
	failOnce   sync.Once
	closing    atomic.Bool
	feedCancel context.CancelFunc
	readyTimer *time.Timer
	wg         sync.WaitGroup // writer, waiter, feeders and an armed ready timer

	// Touched only by offer/offerConfig (under the Session's lock) and by
	// stop (after it).
	dropping     bool
	dropped      int        // packets dropped in the current run
	dropRuns     int        // drop runs started (each one logged)
	totalDropped int        // packets dropped over the encoder's life
	pending      *av.Packet // AVC config still to be queued ahead of the next keyframe
}

// startSessionEncoder starts ffmpeg for the encoded renditions of w.ladder. If
// it cannot, it drops them from the stream, logs why and returns nil.
func startSessionEncoder(ctx context.Context, id string, w encoderWiring) *sessionEncoder {
	encoded := w.ladder[1:]
	disable := func(err error) {
		logger.Errorf(ctx, "livestream %s: encoder unavailable, serving source only: %v", id, err)
		for r := 1; r < len(w.ladder); r++ {
			w.stream.DropRendition(r)
		}
	}

	size := w.transcode.Live.EncoderQueuePackets
	if size < 1 {
		disable(fmt.Errorf("encoderQueuePackets must be positive, got %d", size))
		return nil
	}

	// ffmpeg's lifetime is the session's, ended by stop, not by ctx.
	encCtx := context.WithoutCancel(ctx)
	ff := w.transcode.FFMpegSetting
	args := EncoderArgs(w.src, encoded, ff.Preset, ff.CRF)
	enc, err := StartEncoder(encCtx, id, ff.FFMpegPath, args, len(encoded))
	if err != nil {
		disable(err)
		return nil
	}

	feedCtx, feedCancel := context.WithCancel(encCtx)
	e := &sessionEncoder{
		ctx:        ctx,
		id:         id,
		stream:     w.stream,
		count:      len(encoded),
		enc:        enc,
		queue:      make(chan av.Packet, size),
		halt:       make(chan struct{}),
		feedCancel: feedCancel,
	}

	for i, out := range enc.Outputs() {
		r := i + 1
		name := w.ladder[r].Name
		rec := w.rec
		e.wg.Add(1)
		go func() {
			defer e.wg.Done()
			onAU := func(pts90k, dts90k int64, au [][]byte) { rec.WriteVideo(r, pts90k, dts90k, au) }
			if err := FeedTS(feedCtx, out, w.stream, r, w.clk, onAU); err != nil {
				e.fail(fmt.Errorf("feed rendition %s: %w", name, err))
			}
		}()
	}

	e.wg.Add(2)
	go e.writeLoop(w.headers, w.hasAudio)
	go e.waitLoop()
	return e
}

// offer queues pkt for ffmpeg without blocking and reports whether it was
// queued. A full queue starts a drop run that lasts until a video keyframe
// fits again; a run is logged when it starts and, with its count, when it
// ends. A pending AVC config is queued first; while it cannot be, pkt is
// dropped, so no keyframe reaches ffmpeg ahead of its parameter sets.
func (e *sessionEncoder) offer(pkt av.Packet) bool {
	if e.halted() {
		return false
	}
	if !e.flushPending() {
		e.drop()
		return false
	}

	isKey := pkt.Type == av.H264 && pkt.IsKeyFrame
	if e.dropping && !isKey {
		e.drop()
		return false
	}
	select {
	case e.queue <- pkt:
		if e.dropping {
			e.endDropRun("resumed at a keyframe")
		}
		return true
	default:
		e.drop()
		return false
	}
}

// offerConfig queues a mid-stream AVC config for ffmpeg (ruling R10). It is
// exempt from the drop-until-keyframe rule: when the queue is full it is kept,
// replacing an older pending one, and queued ahead of the next packet that
// fits. It reports whether it was queued at once.
func (e *sessionEncoder) offerConfig(pkt av.Packet) bool {
	if e.halted() {
		return false
	}
	e.pending = &pkt
	if e.flushPending() {
		return true
	}
	e.startDropRun() // the packets behind it wait for it, then for a keyframe
	return false
}

// flushPending queues the pending config, if any, and reports whether none
// is left waiting.
func (e *sessionEncoder) flushPending() bool {
	if e.pending == nil {
		return true
	}
	select {
	case e.queue <- *e.pending:
		e.pending = nil
		return true
	default:
		return false
	}
}

func (e *sessionEncoder) halted() bool {
	select {
	case <-e.halt:
		return true
	default:
		return false
	}
}

// drop counts one packet not given to ffmpeg, starting a run if needed.
func (e *sessionEncoder) drop() {
	e.startDropRun()
	e.dropped++
	e.totalDropped++
}

// startDropRun enters the drop-until-keyframe state and logs it once per run.
func (e *sessionEncoder) startDropRun() {
	if e.dropping {
		return
	}
	e.dropping = true
	e.dropRuns++
	logger.Errorf(e.ctx, "livestream %s: encoder queue full (%d packets), dropping packets for ffmpeg until the next keyframe; encoded renditions stall meanwhile",
		e.id, cap(e.queue))
}

func (e *sessionEncoder) endDropRun(how string) {
	logger.Errorf(e.ctx, "livestream %s: encoder queue drop run over, dropped %d packet(s) for ffmpeg, %s",
		e.id, e.dropped, how)
	e.dropping, e.dropped = false, 0
}

func (e *sessionEncoder) haltEncoder() {
	e.haltOnce.Do(func() { close(e.halt) })
}

// armReadyDeadline fails the encoder if any encoded rendition is not ready
// within d. The Session calls it once, when the source becomes ready.
func (e *sessionEncoder) armReadyDeadline(d time.Duration) {
	select {
	case <-e.halt:
		return
	default:
	}
	e.wg.Add(1)
	e.readyTimer = time.AfterFunc(d, func() {
		defer e.wg.Done()
		for r := 1; r <= e.count; r++ {
			if !e.stream.Ready(r) {
				e.fail(fmt.Errorf("rendition %d not ready %v after the source", r, d))
				return
			}
		}
	})
}

// writeLoop muxes the decoder configs and then every queued packet into
// ffmpeg's stdin as FLV.
func (e *sessionEncoder) writeLoop(headers []av.Packet, hasAudio bool) {
	defer e.wg.Done()

	mux := flv.NewMuxer(e.enc.Input())
	mux.HasVideo, mux.HasAudio = true, hasAudio
	for _, pkt := range headers {
		if err := mux.WritePacket(pkt); err != nil {
			e.fail(fmt.Errorf("write decoder config to ffmpeg: %w", err))
			return
		}
	}
	for {
		select {
		case <-e.halt:
			return
		case pkt := <-e.queue:
			if err := mux.WritePacket(pkt); err != nil {
				e.fail(fmt.Errorf("write to ffmpeg: %w", err))
				return
			}
		}
	}
}

// waitLoop reports ffmpeg exiting on its own.
func (e *sessionEncoder) waitLoop() {
	defer e.wg.Done()
	err := e.enc.Wait()
	if err == nil {
		err = errors.New("exit status 0")
	}
	e.fail(fmt.Errorf("ffmpeg exited: %w", err))
}

// fail handles the first encoder failure while the session is live: it logs
// it, stops feeding the queue, drops every encoded rendition and stops
// ffmpeg. It may block for ffmpeg's stop grace period, so it is never called
// from HandlePacket. Failures during stop are expected and ignored.
func (e *sessionEncoder) fail(err error) {
	if e.closing.Load() {
		return
	}
	e.failOnce.Do(func() {
		logger.Errorf(e.ctx, "livestream %s: encoder failed, encoded renditions dropped, source continues: %v", e.id, err)
		e.haltEncoder()
		for r := 1; r <= e.count; r++ {
			e.stream.DropRendition(r)
		}
		e.feedCancel()
		e.enc.Stop()
	})
}

// stop ends the encoder for Close: it stops feeding, stops the feeders and
// ffmpeg, and waits for every goroutine of the encoder.
func (e *sessionEncoder) stop() {
	e.closing.Store(true)
	e.haltEncoder()
	if e.readyTimer != nil && e.readyTimer.Stop() {
		e.wg.Done() // the timer callback will never run
	}
	e.feedCancel()
	e.enc.Stop()
	e.wg.Wait()
	if e.dropping {
		e.endDropRun("until the session closed")
	}
}
