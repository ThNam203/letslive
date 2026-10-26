package live

import (
	"bytes"
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/nareix/joy5/av"
)

const thumbnailWait = 20 * time.Second

// keyframeAUs returns the access units of the fixture's first n keyframes.
func keyframeAUs(t *testing.T, f streamFixture, n int) [][][]byte {
	t.Helper()
	var aus [][][]byte
	for _, pkt := range f.pkts {
		if pkt.Type != av.H264 || !pkt.IsKeyFrame {
			continue
		}
		_, au, err := f.videoAU(pkt)
		if err != nil {
			t.Fatalf("videoAU: %v", err)
		}
		aus = append(aus, au)
		if len(aus) == n {
			return aus
		}
	}
	t.Fatalf("fixture has fewer than %d keyframes", n)
	return nil
}

func fixtureSource(f streamFixture) SourceInfo {
	return SourceInfo{Width: 640, Height: 360, FPS: 30, SPS: f.sps, PPS: f.pps}
}

func waitForCalls(t *testing.T, st *recordingStorage, n int) []storageCall {
	t.Helper()
	deadline := time.Now().Add(thumbnailWait)
	for time.Now().Before(deadline) {
		if calls := st.Calls(); len(calls) >= n {
			return calls
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("storage got %d calls within %v, want %d", len(st.Calls()), thumbnailWait, n)
	return nil
}

func TestThumbnailer(t *testing.T) {
	requireFFmpeg(t)
	f := loadStreamFixture(t, "src_640x360.flv")
	st := newRecordingStorage(t)

	th := NewThumbnailer(context.Background(), "live1", "ffmpeg", fixtureSource(f), st, time.Hour)
	for _, au := range keyframeAUs(t, f, 3) {
		th.Offer(au)
		time.Sleep(3 * time.Millisecond)
	}
	waitForCalls(t, st, 1)
	th.Close()

	calls := st.Calls()
	if len(calls) != 1 {
		t.Fatalf("AddThumbnail calls = %d, want exactly 1", len(calls))
	}
	call := calls[0]
	if call.Kind != "thumbnail" || call.StreamID != "live1" {
		t.Errorf("call = %+v", call)
	}
	data, err := os.ReadFile(call.SavedPath)
	if err != nil {
		t.Fatalf("read uploaded thumbnail: %v", err)
	}
	if !bytes.HasPrefix(data, []byte{0xFF, 0xD8}) {
		t.Errorf("thumbnail starts with % x, want FF D8", data[:min(len(data), 4)])
	}
	if _, err := os.Stat(call.SourcePath); !os.IsNotExist(err) {
		t.Errorf("temp thumbnail %s still exists after upload (stat err %v)", call.SourcePath, err)
	}
}

func TestThumbnailerOfferNeverBlocks(t *testing.T) {
	requireFFmpeg(t)
	f := loadStreamFixture(t, "src_640x360.flv")
	st := newRecordingStorage(t)
	release := make(chan struct{})
	st.before = func(storageCall) error {
		select {
		case <-release:
		case <-time.After(2 * time.Second):
		}
		return nil
	}

	th := NewThumbnailer(context.Background(), "live1", "ffmpeg", fixtureSource(f), st, time.Nanosecond)
	au := keyframeAUs(t, f, 1)[0]

	// Let the first job reach the slow upload, so the timed Offers meet a busy job.
	th.Offer(au)
	waitForCalls(t, st, 1)

	begin := time.Now()
	for i := 0; i < 100; i++ {
		th.Offer(au)
	}
	if elapsed := time.Since(begin); elapsed > 50*time.Millisecond {
		t.Errorf("100 Offer calls took %v while an upload was in flight, want < 50ms", elapsed)
	}
	close(release)
	th.Close()

	if n := len(st.Calls()); n != 1 {
		t.Errorf("AddThumbnail calls = %d, want 1 (jobs offered while busy are ignored)", n)
	}
}

func TestThumbnailerCloseWaitsForInFlightJob(t *testing.T) {
	requireFFmpeg(t)
	f := loadStreamFixture(t, "src_640x360.flv")
	st := newRecordingStorage(t)
	release := make(chan struct{})
	var finished sync.WaitGroup
	finished.Add(1)
	st.before = func(storageCall) error {
		<-release
		finished.Done()
		return nil
	}

	th := NewThumbnailer(context.Background(), "live1", "ffmpeg", fixtureSource(f), st, time.Hour)
	th.Offer(keyframeAUs(t, f, 1)[0])
	waitForCalls(t, st, 1)

	closed := make(chan struct{})
	go func() {
		th.Close()
		close(closed)
	}()

	select {
	case <-closed:
		t.Fatal("Close returned while the upload was still in flight")
	case <-time.After(100 * time.Millisecond):
	}

	close(release)
	select {
	case <-closed:
	case <-time.After(thumbnailWait):
		t.Fatal("Close did not return after the upload finished")
	}
	finished.Wait()

	th.Close() // second call must be harmless
	th.Offer(keyframeAUs(t, f, 1)[0])
	time.Sleep(50 * time.Millisecond)
	if n := len(st.Calls()); n != 1 {
		t.Errorf("AddThumbnail calls = %d, want 1 (Offer after Close is ignored)", n)
	}
}

func TestThumbnailerSkipsBadInput(t *testing.T) {
	requireFFmpeg(t)
	f := loadStreamFixture(t, "src_640x360.flv")

	t.Run("empty access units", func(t *testing.T) {
		st := newRecordingStorage(t)
		th := NewThumbnailer(context.Background(), "live1", "ffmpeg", fixtureSource(f), st, time.Hour)
		th.Offer(nil)
		th.Offer([][]byte{})
		th.Offer([][]byte{{}})
		th.Offer([][]byte{nil, {}})
		th.Close()
		if n := len(st.Calls()); n != 0 {
			t.Errorf("AddThumbnail calls = %d, want 0", n)
		}
	})

	t.Run("empty access units do not use up the rate limit", func(t *testing.T) {
		st := newRecordingStorage(t)
		th := NewThumbnailer(context.Background(), "live1", "ffmpeg", fixtureSource(f), st, time.Hour)
		th.Offer(nil)
		th.Offer(keyframeAUs(t, f, 1)[0])
		waitForCalls(t, st, 1)
		th.Close()
	})

	t.Run("missing SPS and PPS", func(t *testing.T) {
		st := newRecordingStorage(t)
		src := fixtureSource(f)
		src.SPS, src.PPS = nil, nil
		th := NewThumbnailer(context.Background(), "live1", "ffmpeg", src, st, time.Hour)
		th.Offer(keyframeAUs(t, f, 1)[0])
		th.Close()
		if n := len(st.Calls()); n != 0 {
			t.Errorf("AddThumbnail calls = %d, want 0", n)
		}
	})

	t.Run("undecodable access unit", func(t *testing.T) {
		st := newRecordingStorage(t)
		th := NewThumbnailer(context.Background(), "live1", "ffmpeg", fixtureSource(f), st, time.Hour)
		th.Offer([][]byte{{0x65, 0x00, 0x01, 0x02}})
		th.Close()
		if n := len(st.Calls()); n != 0 {
			t.Errorf("AddThumbnail calls = %d, want 0 (ffmpeg produced no picture)", n)
		}
	})

	t.Run("ffmpeg missing", func(t *testing.T) {
		st := newRecordingStorage(t)
		th := NewThumbnailer(context.Background(), "live1", "/nonexistent/ffmpeg", fixtureSource(f), st, time.Hour)
		th.Offer(keyframeAUs(t, f, 1)[0])
		th.Close()
		if n := len(st.Calls()); n != 0 {
			t.Errorf("AddThumbnail calls = %d, want 0", n)
		}
	})
}

func TestThumbnailerIgnoresOfferAfterClose(t *testing.T) {
	requireFFmpeg(t)
	f := loadStreamFixture(t, "src_640x360.flv")
	st := newRecordingStorage(t)

	// Nothing was accepted before Close and every is tiny, so only the closed
	// flag can reject this offer.
	th := NewThumbnailer(context.Background(), "live1", "ffmpeg", fixtureSource(f), st, time.Nanosecond)
	th.Close()
	th.Offer(keyframeAUs(t, f, 1)[0])
	th.Close()

	time.Sleep(200 * time.Millisecond)
	if n := len(st.Calls()); n != 0 {
		t.Errorf("AddThumbnail calls = %d, want 0 (Offer after Close is ignored)", n)
	}
}

func TestThumbnailerAcceptsAgainAfterEvery(t *testing.T) {
	requireFFmpeg(t)
	f := loadStreamFixture(t, "src_640x360.flv")
	st := newRecordingStorage(t)
	const every = 50 * time.Millisecond

	th := NewThumbnailer(context.Background(), "live1", "ffmpeg", fixtureSource(f), st, every)
	au := keyframeAUs(t, f, 1)[0]

	th.Offer(au)
	waitForCalls(t, st, 1)
	// The job has uploaded; wait until it is no longer "running" and `every` is over.
	time.Sleep(3 * every)
	th.Offer(au)
	waitForCalls(t, st, 2)
	th.Close()

	if n := len(st.Calls()); n != 2 {
		t.Errorf("AddThumbnail calls = %d, want 2", n)
	}
}
