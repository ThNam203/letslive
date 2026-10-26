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

// ErrSessionClosed is returned by HandlePacket after Close.
var ErrSessionClosed = errors.New("session closed")

// SessionDeps is what every Session shares with the rest of the service.
// Registry and Storage are required; Now defaults to time.Now.
type SessionDeps struct {
	Registry  *Registry
	Storage   storage.Storage
	Transcode config.Transcode
	Now       func() time.Time
}

// Session turns one RTMP connection's packets into a live Stream, a VOD
// recording and thumbnails. It buffers the decoder configs, drops media until
// the first H.264 keyframe and builds everything at that keyframe. Nothing
// HandlePacket does waits on ffmpeg, the recorder or the thumbnailer.
//
// HandlePacket and Close may be called from different goroutines.
type Session struct {
	ctx context.Context
	id  string
	d   SessionDeps

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
	stream     *Stream
	rec        *Recorder
	thumbs     *Thumbnailer
	enc        *sessionEncoder // nil when the ladder is source only or ffmpeg failed to start
	srcReady   bool            // the encoded-rendition deadline is armed

	closeOnce sync.Once
	closeDone chan struct{}
	closeErr  error // valid after closeDone is closed
}

// NewSession returns a Session for livestream id. ctx carries the request
// fields attached to log lines; the Session's own work is bounded by Close,
// not by ctx.
func NewSession(ctx context.Context, id string, d SessionDeps) *Session {
	if d.Now == nil {
		d.Now = time.Now
	}
	return &Session{
		ctx:       ctx,
		id:        id,
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
	if s.d.Registry == nil || s.d.Storage == nil {
		return false, fmt.Errorf("livestream %s: session needs a registry and a storage", s.id)
	}
	if s.videoConfig == nil {
		if !s.waitLogged {
			s.waitLogged = true
			logger.Errorf(s.ctx, "livestream %s: keyframe before any AVC config, waiting for one", s.id)
		}
		return false, nil
	}

	tc := s.d.Transcode
	src, err := ProbeAVCConfig(s.videoConfig.Data)
	if err != nil {
		return false, fmt.Errorf("livestream %s: probe AVC config: %w", s.id, err)
	}
	rs, err := BuildLadder(src, tc.FFMpegSetting.Qualities)
	if err != nil {
		return false, fmt.Errorf("livestream %s: build ladder: %w", s.id, err)
	}
	audio := s.probeAudio()

	clk := NewClock(s.d.Now(), key.Time)
	stream, err := NewStream(s.id, rs, audio, tc.Live)
	if err != nil {
		return false, fmt.Errorf("livestream %s: %w", s.id, err)
	}
	if err := s.d.Registry.Add(stream); err != nil {
		stream.Close()
		return false, fmt.Errorf("livestream %s: register stream: %w", s.id, err)
	}

	s.src, s.audio, s.renditions, s.clk, s.stream = src, audio, rs, clk, stream
	s.params = src
	s.rec = NewRecorder(s.ctx, s.id, rs, audio, s.d.Storage, tc.PrivateHLSPath,
		time.Duration(tc.Live.RecordSegmentDurationMs)*time.Millisecond)
	s.thumbs = NewThumbnailer(s.ctx, s.id, tc.FFMpegSetting.FFMpegPath, src, s.d.Storage,
		time.Duration(tc.Live.ThumbnailIntervalMs)*time.Millisecond)
	s.started = true

	if len(rs) > 1 {
		headers := []av.Packet{*s.videoConfig}
		if audio != nil {
			headers = append(headers, *s.audioConfig)
		}
		s.enc = startSessionEncoder(s.ctx, s.id, encoderWiring{
			stream: stream, rec: s.rec, clk: clk, src: src, ladder: rs,
			headers: headers, hasAudio: audio != nil, transcode: tc,
		})
	}

	logger.Infof(s.ctx, "livestream %s: started %dx%d with ladder %v, audio %v",
		s.id, src.Width, src.Height, ladderNamesOf(rs), audio != nil)
	return true, nil
}

// probeAudio parses the buffered AAC config; a bad one makes the stream
// video-only rather than failing it.
func (s *Session) probeAudio() *mpeg4audio.AudioSpecificConfig {
	if s.audioConfig == nil {
		return nil
	}
	audio, err := AudioConfig(*s.audioConfig)
	if err != nil {
		logger.Errorf(s.ctx, "livestream %s: ignoring audio, bad AAC config: %v", s.id, err)
		return nil
	}
	return audio
}

func (s *Session) writeVideo(pkt av.Packet) {
	dts := s.clk.Video90k(pkt.Time)
	if dts < 0 {
		return // earlier than the first keyframe
	}
	if s.enc != nil {
		s.enc.offer(pkt)
	}

	pts, dts, au, nalus, err := s.sourceAU(pkt)
	if err != nil {
		s.logBadInput("skipping video packet at %v: %v", pkt.Time, err)
		return
	}
	if err := s.stream.WriteVideo(0, s.clk.NTP(pts), pts, au); err != nil {
		s.logBadInput("write source video at %v: %v", pkt.Time, err)
	}
	s.rec.WriteVideo(0, pts, dts, au)
	if pkt.IsKeyFrame {
		s.thumbs.Offer(nalus)
	}

	if s.enc != nil && !s.srcReady && s.stream.Ready(0) {
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
	if err := s.stream.WriteAudio(s.clk.NTP(pts90k), s.clk.Audio(pkt.Time, s.audio.SampleRate), aus); err != nil {
		s.logBadInput("write audio at %v: %v", pkt.Time, err)
	}
	s.rec.WriteAudio(pts90k, aus)
}

func (s *Session) logBadInput(format string, args ...any) {
	n, ok := s.badInput.record(time.Now())
	if !ok {
		return
	}
	logger.Errorf(s.ctx, "livestream %s: "+format+" (%d report(s) since the last log)",
		append(append([]any{s.id}, args...), n)...)
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
// idempotent: every call waits for the same shutdown and returns its result,
// which is the recorder's Finish error, if any.
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
		s.d.Registry.Remove(s.id)
	}
	s.mu.Unlock()

	if !started {
		logger.Infof(s.ctx, "livestream %s: closed before the first keyframe, nothing recorded", s.id)
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
		s.enc.stop()
	}
	s.stream.Close()

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
		logger.Errorf(s.ctx, "livestream %s: closed, VOD not written: %v", s.id, finishErr)
		return finishErr
	}
	logger.Infof(s.ctx, "livestream %s: closed", s.id)
	return nil
}
