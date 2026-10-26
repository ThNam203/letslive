package live

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts/codecs"
	"github.com/gofrs/uuid/v5"
	"github.com/nareix/joy5/av"
	"github.com/nareix/joy5/format/flv"

	"sen1or/letslive/transcode/config"
)

// Bounds every wait on ffmpeg or on a Session closing.
const sessionTestTimeout = 30 * time.Second

const sessionMasterFile = "master.m3u8"

var sessionQualities = []config.Quality{
	{Resolution: "1920x1080", MaxBitrate: "5000k", FPS: 30, BufSize: "10000k"},
	{Resolution: "1280x720", MaxBitrate: "2500k", FPS: 30, BufSize: "5000k"},
	{Resolution: "854x480", MaxBitrate: "1000k", FPS: 30, BufSize: "2000k"},
}

// sessionRig is what rtmp.go will hand every Session: one Registry served
// over HTTP, one storage and one transcode config.
type sessionRig struct {
	reg *Registry
	srv *httptest.Server
	st  *recordingStorage
	tc  config.Transcode
}

func newSessionRig(t *testing.T, ffmpegPath string) *sessionRig {
	t.Helper()
	var tc config.Transcode
	tc.PublicHLSPath = t.TempDir()
	tc.PrivateHLSPath = t.TempDir()
	tc.FFMpegSetting.FFMpegPath = ffmpegPath
	tc.FFMpegSetting.MasterFileName = sessionMasterFile
	tc.FFMpegSetting.CRF = 28
	tc.FFMpegSetting.Preset = "ultrafast"
	tc.FFMpegSetting.Qualities = sessionQualities
	tc.Live = config.Live{}.WithDefaults()

	reg := NewRegistry()
	srv := httptest.NewServer(reg.Handler())
	t.Cleanup(srv.Close)
	return &sessionRig{reg: reg, srv: srv, st: newRecordingStorage(t), tc: tc}
}

// newSession starts a Session under a fresh id and closes it at cleanup.
func (r *sessionRig) newSession(t *testing.T) (s *Session, base string) {
	t.Helper()
	id := uuid.Must(uuid.NewV4()).String()
	s = NewSession(context.Background(), id, SessionDeps{
		Registry:  r.reg,
		Storage:   r.st,
		Transcode: r.tc,
		Now:       time.Now,
	})
	t.Cleanup(func() { closeSession(t, s) })
	return s, r.srv.URL + "/" + id
}

func closeSession(t *testing.T, s *Session) error {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), sessionTestTimeout)
	defer cancel()
	err := s.Close(ctx)
	if errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Close did not finish within %v", sessionTestTimeout)
	}
	return err
}

func feedSession(t *testing.T, s *Session, pkts []av.Packet) {
	t.Helper()
	for i, pkt := range pkts {
		if err := s.HandlePacket(pkt); err != nil {
			t.Fatalf("HandlePacket #%d (%v): %v", i, pkt, err)
		}
	}
}

// waitMaster polls until the master playlist is served, then checks it.
func waitMaster(t *testing.T, base string, want, absent []string) string {
	t.Helper()
	deadline := time.Now().Add(sessionTestTimeout)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), testHTTPTimeout)
		resp, err := get(ctx, base+"/index.m3u8")
		cancel()
		if err != nil {
			t.Fatalf("GET master: %v", err)
		}
		if resp.status == http.StatusOK {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("master still %d after %v", resp.status, sessionTestTimeout)
		}
		time.Sleep(50 * time.Millisecond)
	}
	requireMasterLists(t, base, want, absent)
	return mustGet(t, base+"/index.m3u8", http.StatusOK).body
}

// stallScript is an "ffmpeg" that reads nothing and writes nothing. exec keeps
// sleep as the process ffmpeg's pipes belong to, so SIGTERM ends it at once.
func stallScript(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "stall.sh")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexec sleep 30\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// shortThumbnailTimeout keeps the thumbnailer from waiting the full ffmpeg
// timeout on a fake ffmpeg at Close.
func shortThumbnailTimeout(t *testing.T) {
	t.Helper()
	old := thumbnailFFmpegTimeout
	thumbnailFFmpegTimeout = time.Second
	t.Cleanup(func() { thumbnailFFmpegTimeout = old })
}

// Review Focus 1: a source below every configured quality is only repackaged.
func TestSessionSourceOnly(t *testing.T) {
	rig := newSessionRig(t, filepath.Join(t.TempDir(), "no-ffmpeg"))
	s, base := rig.newSession(t)

	feedSession(t, s, readFLV(t, "src_640x360.flv"))

	if got := ladderNamesOf(s.renditions); len(got) != 1 || got[0] != "source" {
		t.Fatalf("ladder = %v, want [source]", got)
	}
	if s.enc != nil {
		t.Error("encoder started for a source-only ladder")
	}
	waitMaster(t, base, []string{"source"}, []string{"1080p", "720p", "480p"})
}

func TestSessionFullLadder(t *testing.T) {
	requireFFmpeg(t)
	rig := newSessionRig(t, ffmpegBinary(t))
	s, base := rig.newSession(t)

	feedSession(t, s, readFLV(t, "src_1280x720.flv"))

	if got := ladderNamesOf(s.renditions); strings.Join(got, ",") != "source,480p" {
		t.Fatalf("ladder = %v, want [source 480p]", got)
	}
	master := waitMaster(t, base, []string{"source", "480p"}, []string{"720p", "1080p"})
	t.Logf("master playlist:\n%s", master)

	if err := closeSession(t, s); err != nil {
		t.Fatalf("Close: %v", err)
	}
	vod, err := os.ReadFile(filepath.Join(rig.tc.PublicHLSPath, "vods", s.id, sessionMasterFile))
	if err != nil {
		t.Fatalf("read VOD master: %v", err)
	}
	for _, want := range []string{"RESOLUTION=1280x720\n0/stream.m3u8", "RESOLUTION=854x480\n1/stream.m3u8"} {
		if !strings.Contains(string(vod), want) {
			t.Errorf("VOD master lacks %q:\n%s", want, vod)
		}
	}
}

// Review Focus 3: ffmpeg missing.
func TestSessionEncoderMissing(t *testing.T) {
	rig := newSessionRig(t, "/nonexistent")
	s, base := rig.newSession(t)

	feedSession(t, s, readFLV(t, "src_1280x720.flv"))

	if !s.stream.Ready(0) {
		t.Error("source not ready")
	}
	waitMaster(t, base, []string{"source"}, []string{"480p"})
}

// Review Focus 3: ffmpeg alive but reading nothing must not slow the source.
// A queue far smaller than the fixture forces the overflow path.
func TestSessionEncoderStalls(t *testing.T) {
	shortThumbnailTimeout(t)
	rig := newSessionRig(t, stallScript(t))
	rig.tc.Live.EncoderQueuePackets = 4
	s, _ := rig.newSession(t)
	pkts := readFLV(t, "src_1280x720.flv")

	elapsed, blocked := feedWithin(t, s, pkts, stallFeedBound)
	if blocked {
		t.Fatalf("HandlePacket blocked: feeding %d packets did not finish within %v", len(pkts), stallFeedBound)
	}
	if elapsed >= time.Second {
		t.Errorf("feeding %d packets took %v, want < 1s", len(pkts), elapsed)
	}
	if s.enc == nil {
		t.Fatal("encoder was not started")
	}
	if s.enc.totalDropped == 0 {
		t.Error("no packet was dropped; the overflow path was not exercised")
	}
	if s.enc.dropRuns == 0 {
		t.Error("no drop run was started (and logged)")
	}
	if !s.stream.Ready(0) {
		t.Error("source not ready")
	}
}

// stallFeedBound is how long feedWithin waits before declaring HandlePacket blocked.
const stallFeedBound = 5 * time.Second

// feedWithin feeds pkts on another goroutine. If that takes longer than bound,
// it reports blocked and drains the encoder queue so the feed (and the
// suite) can finish instead of hanging.
func feedWithin(t *testing.T, s *Session, pkts []av.Packet, bound time.Duration) (elapsed time.Duration, blocked bool) {
	t.Helper()
	done := make(chan error, 1)
	start := time.Now()
	go func() {
		for i, pkt := range pkts {
			if err := s.HandlePacket(pkt); err != nil {
				done <- fmt.Errorf("HandlePacket #%d (%v): %w", i, pkt, err)
				return
			}
		}
		done <- nil
	}()

	var err error
	select {
	case err = <-done:
		elapsed = time.Since(start)
	case <-time.After(bound):
		blocked = true
		elapsed = time.Since(start)
		for drained := false; !drained; {
			select {
			case err = <-done:
				drained = true
			case <-s.enc.queue:
			}
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	return elapsed, blocked
}

// Ruling R6 (c): an encoded rendition not ready in time is dropped and the
// encoder stopped, so the master stops waiting for it.
func TestSessionDropsEncodedNotReadyInTime(t *testing.T) {
	shortThumbnailTimeout(t)
	old := encodedReadyDeadline
	encodedReadyDeadline = 200 * time.Millisecond
	t.Cleanup(func() { encodedReadyDeadline = old })

	rig := newSessionRig(t, stallScript(t))
	s, base := rig.newSession(t)
	feedSession(t, s, readFLV(t, "src_1280x720.flv"))

	waitMaster(t, base, []string{"source"}, []string{"480p"})

	exited := make(chan struct{})
	go func() {
		defer close(exited)
		_ = s.enc.enc.Wait()
	}()
	select {
	case <-exited:
	case <-time.After(sessionTestTimeout):
		t.Fatal("encoder still running after its rendition was dropped")
	}
	select {
	case <-s.enc.halt:
	default:
		t.Error("encoder queue still being fed after the drop")
	}
}

// Review Focus 5.
func TestSessionCloseThenReconnect(t *testing.T) {
	rig := newSessionRig(t, filepath.Join(t.TempDir(), "no-ffmpeg"))
	pkts := readFLV(t, "src_640x360.flv")

	s1, base1 := rig.newSession(t)
	feedSession(t, s1, pkts)
	waitMaster(t, base1, []string{"source"}, nil)

	if err := closeSession(t, s1); err != nil {
		t.Fatalf("Close: %v", err)
	}
	mustGet(t, base1+"/index.m3u8", http.StatusNotFound)

	vodDir := filepath.Join(rig.tc.PublicHLSPath, "vods", s1.id)
	master, err := os.ReadFile(filepath.Join(vodDir, sessionMasterFile))
	if err != nil {
		t.Fatalf("VOD master missing: %v", err)
	}
	if !strings.Contains(string(master), "RESOLUTION=640x360\n0/stream.m3u8") {
		t.Errorf("VOD master:\n%s", master)
	}
	variant, err := os.ReadFile(filepath.Join(vodDir, "0", "stream.m3u8"))
	if err != nil {
		t.Fatalf("VOD variant missing: %v", err)
	}
	// 6s at 4s record segments: one full segment plus the partial last one.
	if n := strings.Count(string(variant), "#EXTINF:"); n != 2 {
		t.Errorf("VOD variant has %d segments, want 2 (incl. the partial one):\n%s", n, variant)
	}

	s2, base2 := rig.newSession(t)
	if s2.id == s1.id {
		t.Fatal("reconnect reused the livestream id")
	}
	feedSession(t, s2, pkts)
	waitMaster(t, base2, []string{"source"}, nil)
	mustGet(t, base1+"/index.m3u8", http.StatusNotFound)
}

func TestSessionHandlePacketAfterClose(t *testing.T) {
	rig := newSessionRig(t, filepath.Join(t.TempDir(), "no-ffmpeg"))
	s, _ := rig.newSession(t)
	pkts := readFLV(t, "src_640x360.flv")
	feedSession(t, s, pkts)

	if err := closeSession(t, s); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := closeSession(t, s); err != nil {
		t.Errorf("second Close: %v", err)
	}
	for _, pkt := range pkts[:5] {
		if err := s.HandlePacket(pkt); !errors.Is(err, ErrSessionClosed) {
			t.Fatalf("HandlePacket after Close = %v, want ErrSessionClosed", err)
		}
	}
}

func TestSessionCloseBeforeFirstKeyframe(t *testing.T) {
	rig := newSessionRig(t, filepath.Join(t.TempDir(), "no-ffmpeg"))
	s, base := rig.newSession(t)

	for _, pkt := range readFLV(t, "src_640x360.flv") {
		if pkt.Type == av.H264 && pkt.IsKeyFrame {
			break
		}
		if err := s.HandlePacket(pkt); err != nil {
			t.Fatalf("HandlePacket: %v", err)
		}
	}
	if s.stream != nil {
		t.Fatal("stream created before the first keyframe")
	}
	mustGet(t, base+"/index.m3u8", http.StatusNotFound)

	if err := closeSession(t, s); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := os.Stat(filepath.Join(rig.tc.PublicHLSPath, "vods", s.id)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("VOD dir for a session that never started: %v", err)
	}
	if calls := rig.st.Calls(); len(calls) != 0 {
		t.Errorf("storage calls = %+v, want none", calls)
	}
}

// Close unregisters at once and returns ctx.Err() when the recorder outlasts
// ctx; the work finishes in the background and a later Close sees its result.
func TestSessionCloseRespectsContext(t *testing.T) {
	rig := newSessionRig(t, filepath.Join(t.TempDir(), "no-ffmpeg"))
	release := make(chan struct{})
	rig.st.before = func(call storageCall) error {
		if call.Kind == "segment" {
			<-release
		}
		return nil
	}
	s, base := rig.newSession(t)
	feedSession(t, s, readFLV(t, "src_640x360.flv"))
	waitMaster(t, base, []string{"source"}, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := s.Close(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Close with a short ctx = %v, want DeadlineExceeded", err)
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("Close returned after %v, want right after ctx expired", elapsed)
	}
	mustGet(t, base+"/index.m3u8", http.StatusNotFound)

	close(release)
	if err := closeSession(t, s); err != nil {
		t.Fatalf("Close after uploads finished: %v", err)
	}
	if _, err := os.Stat(filepath.Join(rig.tc.PublicHLSPath, "vods", s.id, sessionMasterFile)); err != nil {
		t.Errorf("VOD master after background Finish: %v", err)
	}
}

func newQueueTestEncoder(size int) *sessionEncoder {
	return &sessionEncoder{
		ctx:   context.Background(),
		id:    "queue-test",
		queue: make(chan av.Packet, size),
		halt:  make(chan struct{}),
	}
}

func drainQueue(e *sessionEncoder) []av.Packet {
	var got []av.Packet
	for len(e.queue) > 0 {
		got = append(got, <-e.queue)
	}
	return got
}

var (
	queueKey   = av.Packet{Type: av.H264, IsKeyFrame: true}
	queueInter = av.Packet{Type: av.H264}
	queueAudio = av.Packet{Type: av.AAC}
	queueCfg   = av.Packet{Type: av.H264DecoderConfig, Data: []byte{1, 2, 3}}
)

func TestSessionEncoderQueueDropsUntilKeyframe(t *testing.T) {
	e := newQueueTestEncoder(2)

	steps := []struct {
		name     string
		pkt      av.Packet
		drain    bool
		want     bool
		wantRuns int
	}{
		{"inter fits", queueInter, false, true, 0},
		{"audio fits", queueAudio, false, true, 0},
		{"full queue drops inter and starts a run", queueInter, false, false, 1},
		{"full queue drops keyframe", queueKey, false, false, 1},
		{"audio after drain still dropped", queueAudio, true, false, 1},
		{"inter still dropped", queueInter, false, false, 1},
		{"keyframe resumes", queueKey, false, true, 1},
		{"inter after resume", queueInter, false, true, 1},
		{"full again starts a second run", queueInter, false, false, 2},
	}
	for _, st := range steps {
		if st.drain {
			drainQueue(e)
		}
		if got := e.offer(st.pkt); got != st.want {
			t.Fatalf("%s: offer = %v, want %v", st.name, got, st.want)
		}
		if e.dropRuns != st.wantRuns {
			t.Fatalf("%s: dropRuns = %d, want %d", st.name, e.dropRuns, st.wantRuns)
		}
	}
	if e.totalDropped != 5 {
		t.Errorf("totalDropped = %d, want 5", e.totalDropped)
	}

	e.haltEncoder()
	if e.offer(queueKey) {
		t.Error("offer accepted a packet after halt")
	}
}

// Ruling R10: a new AVC config is never dropped by the drop-until-keyframe
// rule and always reaches ffmpeg before the next keyframe.
func TestSessionEncoderQueueConfigBeforeKeyframe(t *testing.T) {
	t.Run("room in the queue", func(t *testing.T) {
		e := newQueueTestEncoder(2)
		if !e.offerConfig(queueCfg) {
			t.Fatal("config not queued although the queue had room")
		}
		if e.dropping || e.dropRuns != 0 {
			t.Errorf("dropping=%v runs=%d, want no drop run", e.dropping, e.dropRuns)
		}
	})

	t.Run("while dropping with room", func(t *testing.T) {
		e := newQueueTestEncoder(2)
		e.offer(queueInter)
		e.offer(queueInter)
		e.offer(queueInter) // starts a run
		drainQueue(e)
		if !e.offerConfig(queueCfg) {
			t.Fatal("config dropped by the drop-until-keyframe rule")
		}
		if e.offer(queueInter) {
			t.Error("inter accepted before the keyframe ended the run")
		}
	})

	t.Run("full queue", func(t *testing.T) {
		e := newQueueTestEncoder(2)
		e.offer(queueInter)
		e.offer(queueAudio)
		if e.offerConfig(queueCfg) {
			t.Fatal("config queued into a full queue")
		}
		if e.pending == nil {
			t.Fatal("config not kept for later delivery")
		}
		if e.offer(queueKey) {
			t.Fatal("keyframe accepted while its config was still pending")
		}
		<-e.queue // room for the config only
		if e.offer(queueKey) {
			t.Fatal("keyframe accepted with no room after the config")
		}
		got := drainQueue(e)
		if len(got) != 2 || got[1].Type != av.H264DecoderConfig {
			t.Fatalf("queue = %v, want [audio, config]", got)
		}
		if !e.offer(queueKey) {
			t.Fatal("keyframe rejected with an empty queue")
		}
		if e.pending != nil || e.dropping {
			t.Errorf("pending=%v dropping=%v after the keyframe, want both cleared", e.pending, e.dropping)
		}
	})
}

// avcConfigWithLevel returns a copy of record whose first SPS has a new
// level_idc: same size, different parameter-set bytes.
func avcConfigWithLevel(t *testing.T, record []byte, level byte) []byte {
	t.Helper()
	const spsLevelOffset = 6 + 2 + 3 // header, SPS length, then NAL header/profile/constraints
	if len(record) <= spsLevelOffset || record[spsLevelOffset] == level {
		t.Fatalf("cannot set level %d in AVC config % x", level, record)
	}
	out := append([]byte(nil), record...)
	out[3], out[spsLevelOffset] = level, level
	return out
}

func decoderConfig(t *testing.T, fixture string) av.Packet {
	t.Helper()
	for _, pkt := range readFLV(t, fixture) {
		if pkt.Type == av.H264DecoderConfig {
			return pkt
		}
	}
	t.Fatalf("%s has no AVC config", fixture)
	return av.Packet{}
}

// keyframeAfter returns the first keyframe at or after d.
func keyframeAfter(t *testing.T, pkts []av.Packet, d time.Duration) av.Packet {
	t.Helper()
	for _, pkt := range pkts {
		if pkt.Type == av.H264 && pkt.IsKeyFrame && pkt.Time >= d {
			return pkt
		}
	}
	t.Fatalf("no keyframe at or after %v", d)
	return av.Packet{}
}

func (s *Session) changeConfig(pkt av.Packet) configChange {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.onVideoConfig(pkt)
}

func (s *Session) nextSourceAU(t *testing.T, pkt av.Packet) [][]byte {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	_, _, au, _, err := s.sourceAU(pkt)
	if err != nil {
		t.Fatalf("sourceAU: %v", err)
	}
	return au
}

func requireParamSets(t *testing.T, au [][]byte, info SourceInfo) {
	t.Helper()
	if len(au) < 2 || !bytes.Equal(au[0], info.SPS) || !bytes.Equal(au[1], info.PPS) {
		t.Fatalf("keyframe AU does not start with the expected SPS/PPS:\n got %x\nwant %x %x", au[:min(2, len(au))], info.SPS, info.PPS)
	}
}

// firstTSKeyframeSPS returns the SPS of the first access unit carrying one in
// the MPEG-TS file at path.
func firstTSKeyframeSPS(t *testing.T, path string) []byte {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	r := &mpegts.Reader{R: f}
	if err := r.Initialize(); err != nil {
		t.Fatalf("mpegts init %s: %v", path, err)
	}
	var sps []byte
	for _, tr := range r.Tracks() {
		if _, ok := tr.Codec.(*codecs.H264); !ok {
			continue
		}
		r.OnDataH264(tr, func(_, _ int64, au [][]byte) error {
			for _, nalu := range au {
				if sps == nil && h264.NALUType(nalu[0]&0x1f) == h264.NALUTypeSPS {
					sps = append([]byte(nil), nalu...)
				}
			}
			return nil
		})
	}
	for sps == nil {
		if err := r.Read(); err != nil {
			t.Fatalf("no SPS in %s: %v", path, err)
		}
	}
	return sps
}

// Ruling R10: same size, new parameter sets.
func TestSessionAVCConfigParamsChange(t *testing.T) {
	rig := newSessionRig(t, filepath.Join(t.TempDir(), "no-ffmpeg"))
	s, base := rig.newSession(t)
	pkts := readFLV(t, "src_640x360.flv")
	old := decoderConfig(t, "src_640x360.flv")
	changed := av.Packet{Type: av.H264DecoderConfig, Data: avcConfigWithLevel(t, old.Data, 0x20)}
	want, err := ProbeAVCConfig(changed.Data)
	if err != nil {
		t.Fatalf("modified config does not parse: %v", err)
	}

	cut := len(pkts) / 2
	feedSession(t, s, pkts[:cut])
	if got := s.changeConfig(changed); got != configParamsChanged {
		t.Fatalf("change = %v, want configParamsChanged", got)
	}
	requireParamSets(t, s.nextSourceAU(t, keyframeAfter(t, pkts[cut:], 0)), want)

	feedSession(t, s, pkts[cut:])
	waitMaster(t, base, []string{"source"}, nil)
	if err := closeSession(t, s); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// The recorder got the same AU: its last segment opens with the new SPS.
	var last string
	for _, c := range rig.st.Calls() {
		if c.Kind == "segment" && c.QualityIndex == 0 {
			last = c.SavedPath
		}
	}
	if last == "" {
		t.Fatal("no source segment uploaded")
	}
	if got := firstTSKeyframeSPS(t, last); !bytes.Equal(got, want.SPS) {
		t.Errorf("recorded SPS after the change = %x, want %x", got, want.SPS)
	}
}

// Ruling R10: a new size keeps the ladder but adopts the new parameter sets.
func TestSessionAVCConfigResolutionChange(t *testing.T) {
	rig := newSessionRig(t, filepath.Join(t.TempDir(), "no-ffmpeg"))
	s, base := rig.newSession(t)
	pkts := readFLV(t, "src_640x360.flv")
	big := decoderConfig(t, "src_1280x720.flv")
	want, err := ProbeAVCConfig(big.Data)
	if err != nil {
		t.Fatal(err)
	}

	cut := len(pkts) / 2
	feedSession(t, s, pkts[:cut])
	if got := s.changeConfig(big); got != configResized {
		t.Fatalf("change = %v, want configResized", got)
	}
	if got := ladderNamesOf(s.renditions); len(got) != 1 || got[0] != "source" {
		t.Errorf("ladder = %v, want it unchanged [source]", got)
	}
	if s.src.Width != 640 || s.src.Height != 360 {
		t.Errorf("ladder source = %dx%d, want the original 640x360", s.src.Width, s.src.Height)
	}
	requireParamSets(t, s.nextSourceAU(t, keyframeAfter(t, pkts[cut:], 0)), want)
	waitMaster(t, base, []string{"source"}, nil)
}

// Ruling R10: an identical repeat changes nothing.
func TestSessionAVCConfigIdenticalRepeat(t *testing.T) {
	rig := newSessionRig(t, filepath.Join(t.TempDir(), "no-ffmpeg"))
	s, _ := rig.newSession(t)
	pkts := readFLV(t, "src_640x360.flv")
	cfg := decoderConfig(t, "src_640x360.flv")
	orig, err := ProbeAVCConfig(cfg.Data)
	if err != nil {
		t.Fatal(err)
	}

	cut := len(pkts) / 2
	feedSession(t, s, pkts[:cut])
	repeat := av.Packet{Type: av.H264DecoderConfig, Data: append([]byte(nil), cfg.Data...)}
	if got := s.changeConfig(repeat); got != configSame {
		t.Fatalf("change = %v, want configSame", got)
	}
	if got := s.changeConfig(av.Packet{Type: av.H264DecoderConfig, Data: []byte{1}}); got != configInvalid {
		t.Errorf("garbage config: change = %v, want configInvalid", got)
	}
	requireParamSets(t, s.nextSourceAU(t, keyframeAfter(t, pkts[cut:], 0)), orig)
}

// dumpScript is an "ffmpeg" that, when run as the encoder, copies its FLV
// input to out; run as the thumbnailer it fails.
func dumpScript(t *testing.T, out string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "dump.sh")
	body := "#!/bin/sh\ncase \"$*\" in\n*filter_complex*) exec cat > '" + out + "' ;;\n*) exit 1 ;;\nesac\n"
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// Ruling R10: ffmpeg receives the new sequence header ahead of the next
// keyframe. (The full-queue case is pinned by
// TestSessionEncoderQueueConfigBeforeKeyframe.)
func TestSessionAVCConfigForwardedToEncoder(t *testing.T) {
	dump := filepath.Join(t.TempDir(), "ffmpeg-input.flv")
	rig := newSessionRig(t, dumpScript(t, dump))
	s, _ := rig.newSession(t)
	pkts := readFLV(t, "src_1280x720.flv")
	changed := av.Packet{Type: av.H264DecoderConfig, Data: avcConfigWithLevel(t, decoderConfig(t, "src_1280x720.flv").Data, 0x20)}

	cut := len(pkts) / 2
	feedSession(t, s, pkts[:cut])
	feedSession(t, s, []av.Packet{changed})
	feedSession(t, s, pkts[cut:])
	if s.enc == nil || s.enc.totalDropped != 0 {
		t.Fatalf("encoder = %v, want running with nothing dropped", s.enc)
	}
	next := keyframeAfter(t, pkts[cut:], 0)

	// Wait until ffmpeg has read past the next keyframe, then check the order.
	deadline := time.Now().Add(sessionTestTimeout)
	for {
		got := ffmpegInput(t, dump)
		cfgAt, keyAt := -1, -1
		for i, pkt := range got {
			switch {
			case pkt.Type == av.H264DecoderConfig && bytes.Equal(pkt.Data, changed.Data):
				cfgAt = i
			case cfgAt < 0 && pkt.Type == av.H264 && pkt.IsKeyFrame && pkt.Time == next.Time:
				t.Fatalf("keyframe at %v reached ffmpeg before the new AVC config", next.Time)
			case pkt.Type == av.H264 && pkt.IsKeyFrame && pkt.Time == next.Time:
				keyAt = i
			}
		}
		if cfgAt >= 0 && keyAt > cfgAt {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("ffmpeg input has config at %d, next keyframe at %d after %v", cfgAt, keyAt, sessionTestTimeout)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// ffmpegInput demuxes what the fake ffmpeg has received so far; a tag cut
// short at the end is ignored.
func ffmpegInput(t *testing.T, path string) []av.Packet {
	t.Helper()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	d := flv.NewDemuxer(bytes.NewReader(data))
	var pkts []av.Packet
	for {
		pkt, err := d.ReadPacket()
		if err != nil {
			return pkts
		}
		pkts = append(pkts, pkt)
	}
}
