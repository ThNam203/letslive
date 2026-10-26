package live

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/nareix/joy5/av"

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
func TestSessionEncoderStalls(t *testing.T) {
	shortThumbnailTimeout(t)
	rig := newSessionRig(t, stallScript(t))
	s, _ := rig.newSession(t)
	pkts := readFLV(t, "src_1280x720.flv")

	start := time.Now()
	feedSession(t, s, pkts)
	if elapsed := time.Since(start); elapsed >= time.Second {
		t.Errorf("feeding %d packets took %v, want < 1s", len(pkts), elapsed)
	}
	if s.enc == nil {
		t.Fatal("encoder was not started")
	}
	if !s.stream.Ready(0) {
		t.Error("source not ready")
	}
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

func TestSessionEncoderQueueDropsUntilKeyframe(t *testing.T) {
	e := &sessionEncoder{
		ctx:   context.Background(),
		id:    "queue-test",
		queue: make(chan av.Packet, 2),
		halt:  make(chan struct{}),
	}
	key := av.Packet{Type: av.H264, IsKeyFrame: true}
	inter := av.Packet{Type: av.H264}
	audio := av.Packet{Type: av.AAC}

	steps := []struct {
		name  string
		pkt   av.Packet
		drain bool
		want  bool
	}{
		{"inter fits", inter, false, true},
		{"audio fits", audio, false, true},
		{"full queue drops inter", inter, false, false},
		{"full queue drops keyframe", key, false, false},
		{"audio after drain still dropped", audio, true, false},
		{"inter still dropped", inter, false, false},
		{"keyframe resumes", key, false, true},
		{"inter after resume", inter, false, true},
	}
	for _, st := range steps {
		if st.drain {
			for len(e.queue) > 0 {
				<-e.queue
			}
		}
		if got := e.offer(st.pkt); got != st.want {
			t.Fatalf("%s: offer = %v, want %v", st.name, got, st.want)
		}
	}
	if e.dropping || e.dropped != 0 {
		t.Errorf("after resume: dropping=%v dropped=%d, want reset", e.dropping, e.dropped)
	}

	e.haltEncoder()
	if e.offer(key) {
		t.Error("offer accepted a packet after halt")
	}
}

func TestSessionResolutionChangeIgnored(t *testing.T) {
	rig := newSessionRig(t, filepath.Join(t.TempDir(), "no-ffmpeg"))
	s, base := rig.newSession(t)

	small := readFLV(t, "src_640x360.flv")
	var bigConfig av.Packet
	for _, pkt := range readFLV(t, "src_1280x720.flv") {
		if pkt.Type == av.H264DecoderConfig {
			bigConfig = pkt
			break
		}
	}
	half := len(small) / 2
	feedSession(t, s, small[:half])
	if err := s.HandlePacket(bigConfig); err != nil {
		t.Fatalf("HandlePacket(new config): %v", err)
	}
	feedSession(t, s, small[half:])

	if s.src.Width != 640 || s.src.Height != 360 {
		t.Errorf("source = %dx%d, want the original 640x360", s.src.Width, s.src.Height)
	}
	waitMaster(t, base, []string{"source"}, nil)
}
