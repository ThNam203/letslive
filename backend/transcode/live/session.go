package live

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
	"github.com/nareix/joy5/av"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/transcode/config"
	"sen1or/letslive/transcode/storage"
)

// sessionLogInterval is the minimum time between two "bad input" log lines
// of one session.
const sessionLogInterval = 5 * time.Second

// encoderEndTimeout bounds how long Close lets ffmpeg drain what it was given.
const encoderEndTimeout = 2 * time.Minute

// ErrSessionClosed is returned by HandlePacket after Close.
var ErrSessionClosed = errors.New("session closed")

// ErrNothingRecorded is returned by Close when the session ended before its
// first keyframe, so no VOD was written.
var ErrNothingRecorded = errors.New("session ended before its first keyframe, nothing recorded")

// SessionDeps is what a Session needs from the rest of the service. Mode
// (from LiveMode or FileMode) and Storage are required; Now defaults to
// time.Now.
type SessionDeps struct {
	Mode      Mode
	Storage   storage.Storage
	Transcode config.Transcode
	Now       func() time.Time
}

// Session turns one source's packets (an RTMP connection or an uploaded
// file) into the Mode's Output, a VOD recording and thumbnails. It buffers the decoder configs, drops media until
// the first H.264 keyframe and builds everything at that keyframe. Nothing
// HandlePacket does waits on ffmpeg, the recorder or the thumbnailer.
//
// HandlePacket and Close may be called from different goroutines.
type Session struct {
	ctx   context.Context
	id    string
	label string // starts every log line, from the Mode
	d     SessionDeps

	// mu serializes HandlePacket with itself and with the start of Close.
	mu          sync.Mutex
	closed      bool
	started     bool
	videoConfig *av.Packet // last H264DecoderConfig before start, then the current one
	audioConfig *av.Packet // last AACDecoderConfig before start
	waitLogged  bool       // "keyframe before AVC config" logged
	badInput    dropLimiter

	// params holds the SPS/PPS (and their size) put ahead of every source
	// keyframe; set at start and replaced by a mid-stream AVC config with
	// different bytes (ruling R10).
	params SourceInfo

	// Set once at start, read-only afterwards. src keeps the size the ladder
	// was built for.
	src        SourceInfo
	audio      *mpeg4audio.AudioSpecificConfig
	renditions []Rendition
	clk        *Clock
	out        Output
	rec        *Recorder
	thumbs     *Thumbnailer
	enc        *sessionEncoder // nil when the ladder is source only or ffmpeg failed to start
	srcReady   bool            // the encoded-rendition deadline is armed

	// lastVideo is the DTS of the latest source video packet relative to the
	// first keyframe; guarded by mu.
	lastVideo time.Duration

	closeOnce sync.Once
	closeDone chan struct{}
	closeErr  error // valid after closeDone is closed
}

// NewSession returns a Session for id (a livestream or VOD id). ctx carries the request
// fields attached to log lines; the Session's own work is bounded by Close,
// not by ctx.
func NewSession(ctx context.Context, id string, d SessionDeps) *Session {
	if d.Now == nil {
		d.Now = time.Now
	}
	return &Session{
		ctx:       ctx,
		id:        id,
		label:     d.Mode.label(id),
		d:         d,
		badInput:  dropLimiter{interval: sessionLogInterval},
		closeDone: make(chan struct{}),
	}
}

// HandlePacket consumes one packet from the RTMP connection. It returns an
// error when the stream cannot be set up at its first keyframe (bad AVC
// config, unusable ladder, duplicate id) and ErrSessionClosed after Close.
// Bad media packets are logged and skipped.
func (s *Session) HandlePacket(pkt av.Packet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrSessionClosed
	}
	if err := s.d.Mode.fail.Err(); err != nil {
		return fmt.Errorf("%s: %w", s.label, err)
	}

	switch pkt.Type {
	case av.H264DecoderConfig:
		s.onVideoConfig(pkt)
	case av.AACDecoderConfig:
		if !s.started {
			s.audioConfig = &pkt
		}
	case av.H264:
		if !s.started {
			if !pkt.IsKeyFrame {
				return nil
			}
			if ok, err := s.start(pkt); !ok {
				return err
			}
		}
		s.writeVideo(pkt)
	case av.AAC:
		if s.started {
			s.writeAudio(pkt)
		}
	}
	return nil
}

// start builds the ladder, stream, recorder, thumbnailer and encoder at the
// first keyframe. ok is false when the keyframe cannot start the session;
// err is nil when that is only because no AVC config has arrived yet.
func (s *Session) start(key av.Packet) (ok bool, err error) {
	if s.d.Mode.open == nil || s.d.Storage == nil {
		return false, fmt.Errorf("%s: session needs a mode and a storage", s.label)
	}
	if s.videoConfig == nil {
		if !s.waitLogged {
			s.waitLogged = true
			logger.Errorf(s.ctx, "%s: keyframe before any AVC config, waiting for one", s.label)
		}
		return false, nil
	}

	tc := s.d.Transcode
	src, err := ProbeAVCConfig(s.videoConfig.Data)
	if err != nil {
		return false, fmt.Errorf("%s: probe AVC config: %w", s.label, err)
	}
	rs, err := BuildLadder(src, tc.FFMpegSetting.Qualities)
	if err != nil {
		return false, fmt.Errorf("%s: build ladder: %w", s.label, err)
	}
	audio := s.probeAudio()

	clk := NewClock(s.d.Now(), key.Time)
	out, err := s.d.Mode.open(s.id, rs, audio)
	if err != nil {
		return false, fmt.Errorf("%s: %w", s.label, err)
	}

	s.src, s.audio, s.renditions, s.clk, s.out = src, audio, rs, clk, out
	s.params = src
	s.rec = NewRecorder(s.ctx, s.label, s.id, rs, audio, s.d.Storage, tc.PrivateHLSPath,
		time.Duration(tc.Live.RecordSegmentDurationMs)*time.Millisecond, s.d.Mode.bp, s.d.Mode.fail)
	s.thumbs = newThumbnailer(s.ctx, s.label, s.id, tc.FFMpegSetting.FFMpegPath, src, s.d.Storage,
		s.d.Mode.thumbs(time.Duration(tc.Live.ThumbnailIntervalMs)*time.Millisecond))
	s.started = true

	if len(rs) > 1 {
		headers := []av.Packet{*s.videoConfig}
		if audio != nil {
			headers = append(headers, *s.audioConfig)
		}
		s.enc = startSessionEncoder(s.ctx, s.label, encoderWiring{
			out: out, rec: s.rec, clk: clk, src: src, ladder: rs,
			headers: headers, hasAudio: audio != nil, transcode: tc,
			bp: s.d.Mode.bp, fail: s.d.Mode.fail, profile: s.d.Mode.profile,
		})
	}

	logger.Infof(s.ctx, "%s: started %dx%d with ladder %v, audio %v",
		s.label, src.Width, src.Height, ladderNamesOf(rs), audio != nil)
	return true, nil
}

// probeAudio parses the buffered AAC config; a bad or missing one makes the
// stream video-only rather than failing it.
func (s *Session) probeAudio() *mpeg4audio.AudioSpecificConfig {
	if s.audioConfig == nil {
		// joy5 only demuxes AAC, so a source sending another audio codec
		// looks the same as one sending no audio.
		logger.Warnf(s.ctx, "%s: no AAC audio before the first keyframe, the stream is video-only (only AAC audio is supported)", s.label)
		return nil
	}
	audio, err := AudioConfig(*s.audioConfig)
	if err != nil {
		logger.Errorf(s.ctx, "%s: ignoring audio, bad AAC config: %v", s.label, err)
		logger.Warnf(s.ctx, "%s: the stream is video-only", s.label)
		return nil
	}
	return audio
}

func (s *Session) writeVideo(pkt av.Packet) {
	dts := s.clk.Video90k(pkt.Time)
	if dts < 0 {
		return // earlier than the first keyframe
	}
	s.lastVideo = max(s.lastVideo, pkt.Time-s.clk.zero)
	if s.enc != nil {
		s.enc.offer(pkt)
	}

	pts, dts, au, nalus, err := s.sourceAU(pkt)
	if err != nil {
		s.logBadInput("skipping video packet at %v: %v", pkt.Time, err)
		s.d.Mode.fail.Report(fmt.Errorf("unusable video packet at %v: %w", pkt.Time, err))
		return
	}
	if err := s.out.WriteVideo(0, s.clk.NTP(pts), pts, au); err != nil {
		s.logBadInput("write source video at %v: %v", pkt.Time, err)
		s.d.Mode.fail.Report(fmt.Errorf("write source video at %v: %w", pkt.Time, err))
	}
	s.rec.WriteVideo(0, pts, dts, au)
	if pkt.IsKeyFrame {
		s.thumbs.Offer(nalus, pkt.Time-s.clk.zero)
	}

	if s.enc != nil && !s.srcReady && s.out.Ready(0) {
		s.srcReady = true
		s.enc.armReadyDeadline(encodedReadyDeadline)
	}
}

// sourceAU converts an RTMP video packet to the source rendition's access
// unit: keyframes get the current parameter sets in front. nalus is the
// packet's own NAL units, for the thumbnailer.
func (s *Session) sourceAU(pkt av.Packet) (pts, dts int64, au, nalus [][]byte, err error) {
	nalus, err = VideoAU(pkt)
	if err != nil {
		return 0, 0, nil, nil, err
	}
	pts = s.clk.Video90k(pkt.Time + pkt.CTime)
	dts = s.clk.Video90k(pkt.Time)
	au = nalus
	if pkt.IsKeyFrame {
		au = withParamSets(s.params.SPS, s.params.PPS, nalus)
	}
	return pts, dts, au, nalus, nil
}

func (s *Session) writeAudio(pkt av.Packet) {
	if s.audio == nil || len(pkt.Data) == 0 {
		return
	}
	pts90k := s.clk.Video90k(pkt.Time)
	if pts90k < 0 {
		return // earlier than the first keyframe
	}
	if s.enc != nil {
		s.enc.offer(pkt)
	}

	aus := [][]byte{pkt.Data}
	if err := s.out.WriteAudio(s.clk.NTP(pts90k), s.clk.Audio(pkt.Time, s.audio.SampleRate), aus); err != nil {
		s.logBadInput("write audio at %v: %v", pkt.Time, err)
		s.d.Mode.fail.Report(fmt.Errorf("write audio at %v: %w", pkt.Time, err))
	}
	s.rec.WriteAudio(pts90k, aus)
}

// Duration is how much of the stream was received: from the first source
// keyframe to the latest source video packet, on the stream's own timeline.
// It is 0 until the session starts and keeps its value after Close.
func (s *Session) Duration() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastVideo
}

// ThumbnailURL is the URL of the session's last uploaded thumbnail, or ""
// when none was. It is final once Close has returned.
func (s *Session) ThumbnailURL() string {
	s.mu.Lock()
	started := s.started
	s.mu.Unlock()
	if !started {
		return ""
	}
	return s.thumbs.URL()
}

func (s *Session) logBadInput(format string, args ...any) {
	n, ok := s.badInput.record(time.Now())
	if !ok {
		return
	}
	logger.Errorf(s.ctx, "%s: "+format+" (%d report(s) since the last log)",
		append(append([]any{s.label}, args...), n)...)
}

// withParamSets returns a new access unit with sps and pps ahead of nalus,
// unless nalus already carries an SPS. RTMP keyframes usually lack them, and
// gohlslib and the recorder both need them in-band. The parameter sets are
// copied so no consumer shares them with the session.
func withParamSets(sps, pps []byte, nalus [][]byte) [][]byte {
	for _, nalu := range nalus {
		if len(nalu) > 0 && h264.NALUType(nalu[0]&0x1f) == h264.NALUTypeSPS {
			return nalus
		}
	}
	au := make([][]byte, 0, len(nalus)+2)
	au = append(au, bytes.Clone(sps), bytes.Clone(pps))
	return append(au, nalus...)
}

func ladderNamesOf(rs []Rendition) []string {
	names := make([]string, len(rs))
	for i, r := range rs {
		names[i] = r.Name
	}
	return names
}

// Close ends the session: the stream is unregistered at once (viewers get
// 404), the encoder is stopped, the stream closed, and then the recording is
// finished and the thumbnailer closed concurrently. If ctx ends first Close
// returns ctx.Err() while that work continues in the background, still
// bounded by the recorder's and thumbnailer's own timeouts. Close is
// idempotent: every call waits for the same shutdown and returns its result:
// the recorder's Finish error, ErrNothingRecorded if the session never
// started, or nil once the VOD is written.
func (s *Session) Close(ctx context.Context) error {
	s.closeOnce.Do(s.beginClose)
	select {
	case <-s.closeDone:
		return s.closeErr
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Session) beginClose() {
	s.mu.Lock()
	s.closed = true
	started := s.started
	if started {
		s.out.Unpublish()
	}
	s.mu.Unlock()

	if !started {
		logger.Infof(s.ctx, "%s: closed before the first keyframe, nothing recorded", s.label)
		s.closeErr = ErrNothingRecorded
		close(s.closeDone)
		return
	}
	go func() {
		s.closeErr = s.shutdown()
		close(s.closeDone)
	}()
}

func (s *Session) shutdown() error {
	if s.enc != nil {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(s.ctx), encoderEndTimeout)
		s.d.Mode.endEncoder(s.enc, ctx)
		cancel()
	}
	s.out.Close()

	var finishErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		finishErr = s.rec.Finish(s.d.Transcode.PublicHLSPath, s.d.Transcode.FFMpegSetting.MasterFileName)
	}()
	go func() {
		defer wg.Done()
		s.thumbs.Close()
	}()
	wg.Wait()

	if finishErr != nil {
		logger.Errorf(s.ctx, "%s: closed, VOD not written: %v", s.label, finishErr)
		return finishErr
	}
	logger.Infof(s.ctx, "%s: closed", s.label)
	return nil
}
