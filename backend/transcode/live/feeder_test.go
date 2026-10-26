package live

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
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

func TestFeedTSReturnsNilWhenRenditionDropped(t *testing.T) {
	requireFFmpeg(t)

	ctx, cancel := context.WithTimeout(context.Background(), testEncoderTimeout)
	defer cancel()

	f := loadStreamFixture(t, "src_1280x720.flv")
	s := newTestStream(t, []string{"source", "480p"}, f.audio, testLiveConfig())

	enc, written := startFixtureEncoder(t, ctx)

	first := make(chan struct{})
	fed := make(chan error, 1)
	go func() {
		var once bool
		fed <- FeedTS(ctx, enc.Outputs()[0], s, 1, f.clock, func(int64, int64, [][]byte) {
			if !once {
				once = true
				close(first)
			}
		})
	}()

	select {
	case <-first:
	case <-ctx.Done():
		t.Fatal("no access unit reached the stream")
	}
	s.DropRendition(1)

	select {
	case err := <-fed:
		if err != nil {
			t.Errorf("FeedTS after DropRendition = %v, want nil", err)
		}
	case <-ctx.Done():
		t.Fatal("FeedTS kept running after its rendition was dropped")
	}

	enc.Stop() // unblocks the FLV writer if ffmpeg stopped reading
	<-written
}

func TestFeedTSStopsOnClosedStream(t *testing.T) {
	requireFFmpeg(t)

	ctx, cancel := context.WithTimeout(context.Background(), testEncoderTimeout)
	defer cancel()

	f := loadStreamFixture(t, "src_1280x720.flv")
	s := newTestStream(t, []string{"source", "480p"}, f.audio, testLiveConfig())
	enc, written := startFixtureEncoder(t, ctx)

	first := make(chan struct{})
	fed := make(chan error, 1)
	go func() {
		var once bool
		fed <- FeedTS(ctx, enc.Outputs()[0], s, 1, f.clock, func(int64, int64, [][]byte) {
			if !once {
				once = true
				close(first)
			}
		})
	}()
	select {
	case <-first:
	case <-ctx.Done():
		t.Fatal("no access unit reached the stream")
	}
	s.Close()

	select {
	case err := <-fed:
		if err != nil {
			t.Errorf("FeedTS after Close = %v, want nil", err)
		}
	case <-ctx.Done():
		t.Fatal("FeedTS kept running after the stream was closed")
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
