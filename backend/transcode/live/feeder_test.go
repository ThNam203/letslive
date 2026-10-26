package live

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts/codecs"
	"github.com/nareix/joy5/format/flv"
)

// startFixtureEncoder starts a single-480p ffmpeg and feeds the whole 720p
// fixture to it from a goroutine. The returned channel yields the write
// error (nil on success) once the writer is done; the input is closed after a
// clean write so ffmpeg flushes and exits.
func startFixtureEncoder(t *testing.T, ctx context.Context) (*Encoder, <-chan error) {
	t.Helper()

	args := EncoderArgs(landscape30(), []Rendition{encodedRendition(480, 30)}, "veryfast", 23)
	enc, err := StartEncoder(ctx, ffmpegBinary(t), args, 1)
	if err != nil {
		t.Fatalf("StartEncoder: %v", err)
	}
	t.Cleanup(enc.Stop)

	pkts := readFLV(t, "src_1280x720.flv")
	written := make(chan error, 1)
	go func() {
		mux := flv.NewMuxer(enc.Input())
		for _, pkt := range pkts {
			if err := mux.WritePacket(pkt); err != nil {
				written <- err
				return
			}
		}
		written <- enc.Input().Close()
	}()
	return enc, written
}

func TestFeedTS(t *testing.T) {
	requireFFmpeg(t)

	ctx, cancel := context.WithTimeout(context.Background(), testEncoderTimeout)
	defer cancel()

	f := loadStreamFixture(t, "src_1280x720.flv")
	s := newTestStream(t, []string{"source", "480p"}, f.audio, testLiveConfig())
	_, base := serveStream(t, s)

	enc, written := startFixtureEncoder(t, ctx)

	var aus, firstDTS, firstPTS int64 = 0, -1, -1
	var lastPTS int64
	fed := make(chan error, 1)
	go func() {
		fed <- FeedTS(ctx, enc.Outputs()[0], s, 1, f.clock, func(pts90k, dts90k int64, au [][]byte) {
			if firstPTS < 0 {
				firstPTS, firstDTS = pts90k, dts90k
			}
			lastPTS = pts90k
			aus++
		})
	}()

	f.feed(t, s, 1, f.pkts) // source rendition and audio, like a session does

	if err := <-written; err != nil {
		t.Fatalf("write FLV to ffmpeg: %v", err)
	}
	select {
	case err := <-fed:
		if err != nil {
			t.Fatalf("FeedTS: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("FeedTS did not return after ffmpeg finished")
	}
	if err := enc.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}

	master := mustGet(t, base+"/index.m3u8", http.StatusOK)
	t.Logf("merged master:\n%s", master.body)
	src, enc480 := strings.Index(master.body, "source/"), strings.Index(master.body, "480p/")
	if src < 0 || enc480 < 0 || src > enc480 {
		t.Errorf("master must list source/ then 480p/ (indexes %d, %d):\n%s", src, enc480, master.body)
	}
	requireMasterLists(t, base, []string{"source", "480p"}, nil)

	media := mustGet(t, base+"/480p/video1_stream.m3u8", http.StatusOK)
	if !strings.Contains(media.body, "#EXT-X-PART:") {
		t.Errorf("480p media playlist has no parts:\n%s", media.body)
	}

	if aus == 0 {
		t.Fatal("onAU never called")
	}
	if firstDTS > firstPTS {
		t.Errorf("first dts %d > pts %d", firstDTS, firstPTS)
	}
	// The first source keyframe is at clock zero; the encoded timeline must start near it.
	if firstPTS < -3000 || firstPTS > 9000 {
		t.Errorf("first 480p pts = %d, want within [-3000, 9000] of the source keyframe (0)", firstPTS)
	}
	if lastPTS <= firstPTS {
		t.Errorf("pts did not advance: first %d last %d", firstPTS, lastPTS)
	}
}

// feedAndStopAtFirstAU runs FeedTS over a full ffmpeg output and calls stop
// from inside the first onAU, so the stream is dropped or closed after exactly
// one access unit. It returns FeedTS's result and how many onAU calls it saw.
// The encoder keeps producing (the whole fixture goes in), so a FeedTS that
// ignored the terminal error would run on and count many more.
func feedAndStopAtFirstAU(t *testing.T, stop func(s *Stream)) (err error, calls int) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), testEncoderTimeout)
	defer cancel()

	f := loadStreamFixture(t, "src_1280x720.flv")
	s := newTestStream(t, []string{"source", "480p"}, f.audio, testLiveConfig())
	enc, written := startFixtureEncoder(t, ctx)

	fed := make(chan error, 1)
	go func() {
		fed <- FeedTS(ctx, enc.Outputs()[0], s, 1, f.clock, func(int64, int64, [][]byte) {
			calls++
			if calls == 1 {
				stop(s)
			}
		})
	}()

	select {
	case err = <-fed:
	case <-ctx.Done():
		t.Fatal("FeedTS did not return")
	}
	enc.Stop() // unblocks the FLV writer if ffmpeg stopped reading
	<-written
	return err, calls
}

func TestFeedTSReturnsNilWhenRenditionDropped(t *testing.T) {
	requireFFmpeg(t)

	err, calls := feedAndStopAtFirstAU(t, func(s *Stream) { s.DropRendition(1) })
	if err != nil {
		t.Errorf("FeedTS after DropRendition = %v, want nil", err)
	}
	if calls != 1 {
		t.Errorf("onAU called %d times, want 1: FeedTS must stop at the first write to a dropped rendition", calls)
	}
}

func TestFeedTSStopsOnClosedStream(t *testing.T) {
	requireFFmpeg(t)

	err, calls := feedAndStopAtFirstAU(t, func(s *Stream) { s.Close() })
	if err != nil {
		t.Errorf("FeedTS after Close = %v, want nil", err)
	}
	if calls != 1 {
		t.Errorf("onAU called %d times, want 1: FeedTS must stop at the first write to a closed stream", calls)
	}
}

// countingReader tracks Read calls in flight so a test can tell that the
// goroutine behind FeedTS has left the reader.
type countingReader struct {
	io.ReadCloser
	active atomic.Int32
}

func (c *countingReader) Read(p []byte) (int, error) {
	c.active.Add(1)
	defer c.active.Add(-1)
	return c.ReadCloser.Read(p)
}

// A cancel arriving while an access unit is being handled must not let that
// handling outlive FeedTS: a caller finalising its recorder right after
// FeedTS returns would otherwise see a late onAU.
func TestFeedTSWaitsForReaderOnContextCancel(t *testing.T) {
	requireFFmpeg(t)

	outer, stop := context.WithTimeout(context.Background(), testEncoderTimeout)
	defer stop()
	ctx, cancel := context.WithCancel(outer)
	defer cancel()

	f := loadStreamFixture(t, "src_1280x720.flv")
	s := newTestStream(t, []string{"source", "480p"}, f.audio, testLiveConfig())
	enc, written := startFixtureEncoder(t, outer)
	src := &countingReader{ReadCloser: enc.Outputs()[0]}

	const slowAU = 200 * time.Millisecond
	var calls, finished atomic.Int32
	fed := make(chan error, 1)
	go func() {
		fed <- FeedTS(ctx, src, s, 1, f.clock, func(int64, int64, [][]byte) {
			if calls.Add(1) == 3 {
				cancel()
				time.Sleep(slowAU) // still inside the callback when ctx is done
				finished.Store(1)
			}
		})
	}()

	select {
	case err := <-fed:
		if err != nil {
			t.Errorf("FeedTS after cancel = %v, want nil", err)
		}
	case <-outer.Done():
		t.Fatal("FeedTS did not return")
	}

	if finished.Load() != 1 {
		t.Error("FeedTS returned while onAU was still running")
	}
	if n := src.active.Load(); n != 0 {
		t.Errorf("%d reads still in flight after FeedTS returned", n)
	}
	atReturn := calls.Load()
	time.Sleep(300 * time.Millisecond)
	if after := calls.Load(); after != atReturn {
		t.Errorf("onAU called %d more times after FeedTS returned", after-atReturn)
	}

	enc.Stop()
	<-written
}

func TestFeedTSReturnsPromptlyOnContextCancel(t *testing.T) {
	requireFFmpeg(t)

	outer, stop := context.WithTimeout(context.Background(), testEncoderTimeout)
	defer stop()

	f := loadStreamFixture(t, "src_1280x720.flv")
	s := newTestStream(t, []string{"source", "480p"}, f.audio, testLiveConfig())

	// ffmpeg gets no input, so its output pipe stays open and silent.
	args := EncoderArgs(landscape30(), []Rendition{encodedRendition(480, 30)}, "veryfast", 23)
	enc, err := StartEncoder(outer, ffmpegBinary(t), args, 1)
	if err != nil {
		t.Fatalf("StartEncoder: %v", err)
	}
	t.Cleanup(enc.Stop)

	ctx, cancel := context.WithCancel(outer)
	fed := make(chan error, 1)
	go func() { fed <- FeedTS(ctx, enc.Outputs()[0], s, 1, f.clock, nil) }()

	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case err := <-fed:
		if err != nil {
			t.Errorf("FeedTS after cancel = %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("FeedTS did not return promptly after ctx cancel")
	}
}

func TestFeedTSNoH264Track(t *testing.T) {
	var ts bytes.Buffer
	track := &mpegts.Track{Codec: &codecs.MPEG4Audio{Config: mpeg4audio.Config{
		Type: mpeg4audio.ObjectTypeAACLC, SampleRate: 48000, ChannelCount: 2,
	}}}
	w := &mpegts.Writer{W: &ts, Tracks: []*mpegts.Track{track}}
	if err := w.Initialize(); err != nil {
		t.Fatalf("mpegts writer: %v", err)
	}
	for i := range 20 {
		if err := w.WriteMPEG4Audio(track, int64(i)*1920, [][]byte{{0x21, 0x10, 0x04, 0x60}}); err != nil {
			t.Fatalf("write audio: %v", err)
		}
	}

	f := loadStreamFixture(t, "src_1280x720.flv")
	s := newTestStream(t, []string{"source", "480p"}, f.audio, testLiveConfig())

	ctx, cancel := context.WithTimeout(context.Background(), testEncoderTimeout)
	defer cancel()
	err := FeedTS(ctx, &ts, s, 1, f.clock, nil)
	if err == nil {
		t.Fatal("FeedTS on an audio-only TS = nil, want error")
	}
	if !strings.Contains(err.Error(), "H.264") {
		t.Errorf("error = %v, want it to mention H.264", err)
	}
}

func TestFeedTSBadRendition(t *testing.T) {
	f := loadStreamFixture(t, "src_1280x720.flv")
	s := newTestStream(t, []string{"source"}, f.audio, testLiveConfig())
	err := FeedTS(context.Background(), io.LimitReader(bytes.NewReader(nil), 0), s, 5, f.clock, nil)
	if err == nil {
		t.Fatal("FeedTS with rendition out of range = nil, want error")
	}
}
