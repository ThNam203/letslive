package live

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
	"github.com/gofrs/uuid/v5"
)

// testMode is a Mode with the given strategies; the rest are live's.
func testMode(t *testing.T, open func(id string, rs []Rendition) Output, bp Backpressure, f Failures) Mode {
	t.Helper()
	m := mustLiveMode(t, NewRegistry())
	m.noun = "test"
	if open != nil {
		m.open = func(id string, rs []Rendition, _ *mpeg4audio.AudioSpecificConfig) (Output, error) {
			return open(id, rs), nil
		}
	}
	m.bp, m.fail = bp, f
	return m
}

func openNothing(_ string, rs []Rendition) Output { return newNoOutput(rs) }

// newFileishSession starts a Session that waits on full queues and fails on
// the first reported error, with nobody watching live.
func newFileishSession(t *testing.T, rig *sessionRig) (*Session, *firstError) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	fail := &firstError{cancel: cancel}
	id := uuid.Must(uuid.NewV4()).String()
	s := NewSession(context.Background(), id, SessionDeps{
		Mode:      testMode(t, openNothing, waitUntil{ctx: ctx}, fail),
		Storage:   rig.st,
		Transcode: rig.tc,
		Now:       time.Now,
	})
	t.Cleanup(func() { closeSession(t, s) })
	return s, fail
}

func TestSendDropWhenFull(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 1
	start := time.Now()
	if send(dropWhenFull{}, ch, 2) {
		t.Fatal("send to a full queue = true, want false")
	}
	if d := time.Since(start); d > 50*time.Millisecond {
		t.Errorf("send took %v, want it not to wait", d)
	}
}

func TestSendWaitUntilRoom(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 1
	go func() {
		time.Sleep(50 * time.Millisecond)
		<-ch
	}()
	if !send(waitUntil{ctx: context.Background()}, ch, 2) {
		t.Fatal("send = false, want it to wait for room")
	}
	if got := <-ch; got != 2 {
		t.Errorf("queued %d, want 2", got)
	}
}

func TestSendWaitGivesUpOnCancel(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 1
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(20*time.Millisecond, cancel)
	start := time.Now()
	if send(waitUntil{ctx: ctx}, ch, 2) {
		t.Fatal("send = true after cancel, want false")
	}
	if d := time.Since(start); d > 120*time.Millisecond {
		t.Errorf("send gave up after %v, want within ~100ms of cancel", d)
	}
}

func TestFirstErrorKeepsFirstAndCancels(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	f := &firstError{cancel: cancel}
	if f.Err() != nil {
		t.Fatalf("Err before any report = %v", f.Err())
	}
	first := errors.New("first")
	f.Report(first)
	f.Report(errors.New("second"))
	if !errors.Is(f.Err(), first) {
		t.Errorf("Err = %v, want the first error", f.Err())
	}
	if ctx.Err() == nil {
		t.Error("Report did not cancel")
	}
}

func TestLogOnlyNeverFails(t *testing.T) {
	var f logOnly
	f.Report(errors.New("x"))
	if f.Err() != nil {
		t.Errorf("Err = %v, want nil", f.Err())
	}
}

// Review Focus 1: a file is fed faster than ffmpeg encodes; nothing may be
// dropped and every rendition must be written.
func TestFileModeNeverDrops(t *testing.T) {
	requireFFmpeg(t)
	rig := newSessionRig(t, ffmpegBinary(t))
	rig.tc.Live.EncoderQueuePackets = 2
	s, fail := newFileishSession(t, rig)

	feedSession(t, s, readFLV(t, "src_1280x720.flv"))
	if err := closeSession(t, s); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := fail.Err(); err != nil {
		t.Fatalf("failure reported: %v", err)
	}
	if n := s.rec.Dropped(); n != 0 {
		t.Errorf("recorder dropped %d item(s), want 0", n)
	}
	if s.enc.totalDropped != 0 {
		t.Errorf("encoder queue dropped %d packet(s), want 0", s.enc.totalDropped)
	}
	for _, r := range []string{"0", "1"} {
		if _, err := os.Stat(filepath.Join(rig.tc.PublicHLSPath, "vods", s.id, r, "stream.m3u8")); err != nil {
			t.Errorf("rendition %s playlist: %v", r, err)
		}
	}
}

func TestFailuresReportedOnEncoderFailure(t *testing.T) {
	rig := newSessionRig(t, filepath.Join(t.TempDir(), "no-ffmpeg"))
	s, fail := newFileishSession(t, rig)

	pkts := readFLV(t, "src_1280x720.flv")
	for _, pkt := range pkts {
		if err := s.HandlePacket(pkt); err != nil {
			break
		}
	}
	if fail.Err() == nil || !strings.Contains(fail.Err().Error(), "encoder") {
		t.Errorf("Err = %v, want the encoder failure", fail.Err())
	}
}

func TestFailuresReportedOnSegmentUploadFailure(t *testing.T) {
	rig := newSessionRig(t, filepath.Join(t.TempDir(), "no-ffmpeg"))
	rig.st.before = func(c storageCall) error {
		if c.Kind == "segment" {
			return errors.New("minio down")
		}
		return nil
	}
	s, fail := newFileishSession(t, rig)

	for _, pkt := range readFLV(t, "src_640x360.flv") {
		if err := s.HandlePacket(pkt); err != nil {
			break
		}
	}
	closeSession(t, s)
	if fail.Err() == nil || !strings.Contains(fail.Err().Error(), "minio down") {
		t.Errorf("Err = %v, want the upload failure", fail.Err())
	}
}

func TestFailuresReportedOnResize(t *testing.T) {
	rig := newSessionRig(t, filepath.Join(t.TempDir(), "no-ffmpeg"))
	s, fail := newFileishSession(t, rig)
	pkts := readFLV(t, "src_640x360.flv")
	feedSession(t, s, pkts[:len(pkts)/2])

	if got := s.changeConfig(decoderConfig(t, "src_1280x720.flv")); got != configResized {
		t.Fatalf("change = %v, want configResized", got)
	}
	if fail.Err() == nil || !strings.Contains(fail.Err().Error(), "resolution changed") {
		t.Errorf("Err = %v, want the resolution change", fail.Err())
	}
}

func TestHandlePacketReturnsReportedFailure(t *testing.T) {
	rig := newSessionRig(t, filepath.Join(t.TempDir(), "no-ffmpeg"))
	s, fail := newFileishSession(t, rig)
	pkts := readFLV(t, "src_640x360.flv")
	feedSession(t, s, pkts[:len(pkts)/2])

	fail.Report(errors.New("boom"))
	if err := s.HandlePacket(pkts[len(pkts)/2]); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("HandlePacket after a failure = %v, want it to return boom", err)
	}
}
