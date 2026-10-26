package live

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bluenviron/gohlslib/v2"
	"github.com/bluenviron/gohlslib/v2/pkg/codecs"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/transcode/config"
)

const (
	hlsContentType     = "application/vnd.apple.mpegurl"
	masterPlaylistName = "index.m3u8"
	// Same policy gohlslib uses for its own multivariant playlist.
	masterCacheControl = "public, max-age=30"
	notReadyCacheCtl   = "no-store"
	// Upper bound for reading one muxer's index.m3u8 while building the master.
	// The readiness gate makes that read immediate; this only guards against
	// the gate and gohlslib disagreeing.
	masterBuildTimeout = 2 * time.Second
	// gohlslib v2.4.7 adds this to every fMP4 DTS (muxer.go fmp4StartDTS).
	gohlslibFMP4StartDTS = 10 * time.Second
)

// ErrStreamClosed is returned by writes after Close.
var ErrStreamClosed = errors.New("stream closed")

// ErrRenditionDropped is returned by WriteVideo for a rendition removed with
// DropRendition.
var ErrRenditionDropped = errors.New("rendition dropped")

// Stream is one livestream's LL-HLS output: a gohlslib Muxer per rendition
// (gohlslib supports one video track per Muxer) and a master playlist that
// merges the muxers' own multivariant playlists.
type Stream struct {
	id         string
	renditions []*renditionMuxer
	byName     map[string]*renditionMuxer
	hasAudio   bool
	closed     atomic.Bool

	masterMu sync.Mutex
	master   []byte // built once every rendition is ready
}

type renditionMuxer struct {
	name  string
	muxer *gohlslib.Muxer
	video *gohlslib.Track
	audio *gohlslib.Track // nil for a video-only stream

	// mu serializes writes and Close: gohlslib's write path is not locked,
	// and video (session or ffmpeg feeder) and audio (session) arrive from
	// different goroutines.
	mu     sync.Mutex
	closed bool
	gate   segmentGate
	ready  atomic.Bool
	// dropped is set under mu, read lock-free by routing and the master build.
	dropped atomic.Bool
}

// NewStream starts one LL-HLS muxer per rendition. When audio is non-nil every
// muxer carries an H.264 and an MPEG-4 audio track, otherwise video only.
func NewStream(id string, rs []Rendition, audio *mpeg4audio.AudioSpecificConfig, cfg config.Live) (*Stream, error) {
	if len(rs) == 0 {
		return nil, errors.New("stream needs at least one rendition")
	}
	if audio != nil && audio.SampleRate <= 0 {
		return nil, fmt.Errorf("invalid audio sample rate %d", audio.SampleRate)
	}
	if cfg.PartMinDurationMs <= 0 || cfg.SegmentMinDurationMs <= 0 {
		return nil, fmt.Errorf("part (%dms) and segment (%dms) durations must be positive",
			cfg.PartMinDurationMs, cfg.SegmentMinDurationMs)
	}

	s := &Stream{
		id:       id,
		byName:   make(map[string]*renditionMuxer, len(rs)),
		hasAudio: audio != nil,
	}
	for _, r := range rs {
		if err := s.addRendition(r.Name, audio, cfg); err != nil {
			s.Close()
			return nil, err
		}
	}
	return s, nil
}

func (s *Stream) addRendition(name string, audio *mpeg4audio.AudioSpecificConfig, cfg config.Live) error {
	if !renditionNameRe.MatchString(name) {
		return fmt.Errorf("rendition name %q must match %s", name, renditionNameRe)
	}
	if _, dup := s.byName[name]; dup {
		return fmt.Errorf("duplicate rendition %q", name)
	}

	m := &renditionMuxer{
		name:  name,
		video: &gohlslib.Track{Codec: &codecs.H264{}, ClockRate: videoClockRate},
		gate:  segmentGate{minDuration: time.Duration(cfg.SegmentMinDurationMs) * time.Millisecond},
	}
	tracks := []*gohlslib.Track{m.video}
	if audio != nil {
		m.audio = &gohlslib.Track{Codec: &codecs.MPEG4Audio{Config: *audio}, ClockRate: audio.SampleRate}
		tracks = append(tracks, m.audio)
	}

	m.muxer = &gohlslib.Muxer{
		Tracks:             tracks,
		Variant:            gohlslib.MuxerVariantLowLatency,
		SegmentCount:       cfg.SegmentCount,
		SegmentMinDuration: m.gate.minDuration,
		PartMinDuration:    time.Duration(cfg.PartMinDurationMs) * time.Millisecond,
		OnEncodeError: func(err error) {
			logger.Errorf(context.Background(), "livestream %s rendition %s: hls muxer: %v", s.id, name, err)
		},
	}
	if err := m.muxer.Start(); err != nil {
		return fmt.Errorf("start %s muxer: %w", name, err)
	}

	s.renditions = append(s.renditions, m)
	s.byName[name] = m
	return nil
}

// ID is the livestream id the stream is served under.
func (s *Stream) ID() string {
	return s.id
}

// WriteVideo writes one H.264 access unit to rendition r. pts90k is on the
// shared Clock's 90 kHz timeline; keyframes must carry SPS and PPS. gohlslib
// keeps references to the SPS and PPS NAL units, so au must not be reused.
func (s *Stream) WriteVideo(r int, ntp time.Time, pts90k int64, au [][]byte) error {
	if r < 0 || r >= len(s.renditions) {
		return fmt.Errorf("rendition index %d out of range [0, %d)", r, len(s.renditions))
	}
	return s.renditions[r].writeVideo(ntp, pts90k, au)
}

// WriteAudio writes AAC access units to every rendition. pts is in units of
// the audio sample rate. Every rendition is attempted; failures are joined.
// gohlslib keeps the slices until their part is closed, so aus must not be
// reused by the caller.
func (s *Stream) WriteAudio(ntp time.Time, pts int64, aus [][]byte) error {
	if !s.hasAudio {
		return errors.New("stream has no audio track")
	}
	var errs []error
	for _, m := range s.renditions {
		if err := m.writeAudio(ntp, pts, aus); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// ServeHTTP serves "/index.m3u8" (the master playlist) and
// "/{rendition}/{file}" relative to the stream.
func (s *Stream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !allowGET(w, r) {
		return
	}
	s.serve(w, r, strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/"))
}

// serve routes the path segments after the stream id. Muxers route on the
// last element of r.URL.Path, so the request is passed through unchanged.
func (s *Stream) serve(w http.ResponseWriter, r *http.Request, parts []string) {
	if s.closed.Load() {
		http.NotFound(w, r)
		return
	}
	switch {
	case len(parts) == 1 && parts[0] == masterPlaylistName:
		s.serveMaster(w, r)
		return
	case len(parts) == 2 && parts[1] != "":
		if m, ok := s.byName[parts[0]]; ok && !m.dropped.Load() {
			m.serve(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

// DropRendition removes encoded rendition r (index >= 1), e.g. after its
// encoder failed: its muxer is closed, it no longer gates or appears in the
// master playlist, its URLs return 404, WriteVideo to it returns
// ErrRenditionDropped and WriteAudio skips it. The source (r == 0) cannot be
// dropped; that and out-of-range indexes are logged and ignored. Repeated
// calls are no-ops.
func (s *Stream) DropRendition(r int) {
	if r < 1 || r >= len(s.renditions) {
		logger.Errorf(context.Background(), "livestream %s: cannot drop rendition index %d (valid: 1..%d)",
			s.id, r, len(s.renditions)-1)
		return
	}
	if !s.renditions[r].drop() {
		return
	}

	// Taken after the drop is visible, so a build that raced it is discarded.
	s.masterMu.Lock()
	s.master = nil
	s.masterMu.Unlock()
}

// Ready reports whether rendition r has closed its first segment, the point
// from which it can be listed in the master playlist. It is false for a
// dropped rendition and for an index out of range.
func (s *Stream) Ready(r int) bool {
	if r < 0 || r >= len(s.renditions) {
		return false
	}
	m := s.renditions[r]
	return m.ready.Load() && !m.dropped.Load()
}

// Close stops every muxer. Pending blocking playlist requests return 404.
func (s *Stream) Close() {
	if s.closed.Swap(true) {
		return
	}
	for _, m := range s.renditions {
		m.close()
	}
}

func (s *Stream) serveMaster(w http.ResponseWriter, r *http.Request) {
	body, ready, err := s.masterPlaylist(r.Context())
	if err != nil {
		logger.Errorf(r.Context(), "livestream %s: build master playlist: %v", s.id, err)
		w.Header().Set("Cache-Control", notReadyCacheCtl)
		http.Error(w, "master playlist unavailable", http.StatusInternalServerError)
		return
	}
	if !ready {
		w.Header().Set("Cache-Control", notReadyCacheCtl)
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", hlsContentType)
	w.Header().Set("Cache-Control", masterCacheControl)
	if _, err := w.Write(body); err != nil {
		logger.Errorf(r.Context(), "livestream %s: write master playlist: %v", s.id, err)
	}
}

// masterPlaylist returns the cached master, or builds it once every rendition
// has closed its first segment. ready is false before that.
func (s *Stream) masterPlaylist(ctx context.Context) (body []byte, ready bool, err error) {
	s.masterMu.Lock()
	defer s.masterMu.Unlock()

	if s.master != nil {
		return s.master, true, nil
	}
	for _, m := range s.renditions {
		if !m.ready.Load() && !m.dropped.Load() {
			return nil, false, nil
		}
	}

	parts := make([]NamedPlaylist, 0, len(s.renditions))
	for _, m := range s.renditions {
		if m.dropped.Load() {
			continue
		}
		pl, err := m.multivariantPlaylist(ctx)
		if err != nil {
			if m.dropped.Load() { // dropped mid-build: its muxer was closed
				continue
			}
			return nil, false, err
		}
		parts = append(parts, NamedPlaylist{Name: m.name, Body: pl})
	}
	merged, err := MergeMultivariant(parts)
	if err != nil {
		return nil, false, err
	}
	s.master = merged
	return merged, true, nil
}

func (m *renditionMuxer) writeVideo(ntp time.Time, pts90k int64, au [][]byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.dropped.Load() {
		return fmt.Errorf("rendition %s: %w", m.name, ErrRenditionDropped)
	}
	if m.closed {
		return ErrStreamClosed
	}
	err := m.muxer.WriteH264(m.video, ntp, pts90k, au)
	if m.gate.observe(au, pts90k, err == nil) {
		m.ready.Store(true)
	}
	if err != nil {
		return fmt.Errorf("rendition %s: write H264: %w", m.name, err)
	}
	return nil
}

func (m *renditionMuxer) writeAudio(ntp time.Time, pts int64, aus [][]byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.dropped.Load() {
		return nil
	}
	if m.closed {
		return ErrStreamClosed
	}
	if err := m.muxer.WriteMPEG4Audio(m.audio, ntp, pts, aus); err != nil {
		return fmt.Errorf("rendition %s: write MPEG-4 audio: %w", m.name, err)
	}
	return nil
}

// drop marks the rendition dropped and closes its muxer, which also releases
// requests parked on its playlists. It reports whether this call dropped it.
func (m *renditionMuxer) drop() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.dropped.Load() {
		return false
	}
	m.dropped.Store(true)
	if !m.closed {
		m.closed = true
		m.muxer.Close()
	}
	return true
}

func (m *renditionMuxer) close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return
	}
	m.closed = true
	m.muxer.Close()
}

// serve hands the request to the muxer. gohlslib writes nothing for a file it
// does not know, which would be an empty 200, so that case becomes a 404.
func (m *renditionMuxer) serve(w http.ResponseWriter, r *http.Request) {
	tw := &trackingWriter{ResponseWriter: w}
	m.muxer.Handle(tw, r)
	if !tw.wrote {
		http.NotFound(w, r)
	}
}

// multivariantPlaylist reads the muxer's own index.m3u8 in process. gohlslib
// waits on a sync.Cond until the first segment is closed and ignores the
// request context, so this is only called once the gate reports ready. If the
// timeout still fires, the handler goroutine stays parked until the muxer
// closes a segment or is closed.
func (m *renditionMuxer) multivariantPlaylist(ctx context.Context) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, masterBuildTimeout)
	defer cancel()

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/"+masterPlaylistName, nil)
	rec := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		defer close(done)
		m.muxer.Handle(rec, req)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return nil, fmt.Errorf("rendition %s: read %s: %w", m.name, masterPlaylistName, ctx.Err())
	}
	if rec.Code != http.StatusOK {
		return nil, fmt.Errorf("rendition %s: read %s: status %d", m.name, masterPlaylistName, rec.Code)
	}
	return rec.Body.Bytes(), nil
}

// segmentGate tells when gohlslib has closed a rendition's first segment,
// which is when its index.m3u8 stops blocking. It mirrors gohlslib v2.4.7
// (muxer_segmenter.go writeH264 and fmp4WriteSample): access units before the
// first random-access one are dropped; DTS comes from an h264.DTSExtractor fed
// the same access units; the first accepted sample starts the segment; a later
// random-access unit at least minDuration after that start closes it. A second
// keyframe sooner than minDuration therefore does not make the rendition ready.
type segmentGate struct {
	minDuration time.Duration

	started   bool
	dts       h264.DTSExtractor
	haveStart bool
	start     time.Duration
	closed    bool
}

// observe records one access unit; accepted reports whether gohlslib's
// WriteH264 returned nil for it. It returns true once the first segment is closed.
func (g *segmentGate) observe(au [][]byte, pts90k int64, accepted bool) bool {
	if g.closed {
		return true
	}

	randomAccess := h264.IsRandomAccess(au)
	if !g.started {
		if !randomAccess {
			return false
		}
		g.started = true
		g.dts.Initialize()
	}

	// Extract runs even for rejected units, as gohlslib's own extractor did.
	dts, err := g.dts.Extract(au, pts90k)
	if err != nil || !accepted {
		return false
	}

	at := gohlslibDuration(dts)
	if !g.haveStart {
		g.haveStart, g.start = true, at
		return false
	}
	if randomAccess && at-g.start >= g.minDuration {
		g.closed = true
	}
	return g.closed
}

// gohlslibDuration converts a 90 kHz DTS to the duration gohlslib compares
// against SegmentMinDuration, with the same offset and rounding.
func gohlslibDuration(dts90k int64) time.Duration {
	v := dts90k + int64(gohlslibFMP4StartDTS/time.Second)*videoClockRate
	secs := v / videoClockRate
	rem := v % videoClockRate
	return time.Duration(secs)*time.Second + time.Duration(rem)*time.Second/videoClockRate
}

// trackingWriter records whether a handler produced a response.
type trackingWriter struct {
	http.ResponseWriter
	wrote bool
}

func (w *trackingWriter) WriteHeader(code int) {
	w.wrote = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *trackingWriter) Write(b []byte) (int, error) {
	w.wrote = true
	return w.ResponseWriter.Write(b)
}

func (w *trackingWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func allowGET(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodGet {
		return true
	}
	w.Header().Set("Allow", http.MethodGet)
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	return false
}
