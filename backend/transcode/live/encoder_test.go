package live

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts/codecs"
	"github.com/nareix/joy5/av"
	"github.com/nareix/joy5/format/flv"
)

// Bounds every test that waits on a child process, so a stuck ffmpeg fails
// the test (ctx cancel sends SIGTERM, pipes hit EOF) instead of hanging.
const testEncoderTimeout = 30 * time.Second

func encodedRendition(short int, fps int) Rendition {
	return Rendition{
		Name:       fmt.Sprintf("%dp", short),
		ShortSide:  short,
		Encoded:    true,
		MaxBitrate: map[int]string{720: "2500k", 480: "1000k"}[short],
		BufSize:    map[int]string{720: "5000k", 480: "2000k"}[short],
		FPS:        fps,
	}
}

func landscape30() SourceInfo {
	return SourceInfo{Width: 1280, Height: 720, FPS: 30}
}

func TestEncoderArgsLandscape(t *testing.T) {
	got := EncoderArgs(SourceInfo{Width: 1920, Height: 1080, FPS: 30},
		[]Rendition{encodedRendition(720, 30), encodedRendition(480, 30)}, "veryfast", 23)

	want := []string{
		"-hide_banner", "-loglevel", "error", "-probesize", "32", "-analyzeduration", "0", "-f", "flv", "-i", "pipe:0", "-an", "-copyts",
		"-filter_complex", "[0:v]split=2[s0][s1];[s0]scale=-2:720[v0];[s1]scale=-2:480[v1]",
		"-map", "[v0]", "-c:v", "libx264", "-preset", "veryfast", "-tune", "zerolatency", "-pix_fmt", "yuv420p",
		"-crf", "23", "-maxrate", "2500k", "-bufsize", "5000k", "-force_key_frames", "source", "-sc_threshold", "0", "-g", "300", "-r", "30",
		"-muxdelay", "0", "-muxpreload", "0", "-flush_packets", "1", "-f", "mpegts", "pipe:3",
		"-map", "[v1]", "-c:v", "libx264", "-preset", "veryfast", "-tune", "zerolatency", "-pix_fmt", "yuv420p",
		"-crf", "23", "-maxrate", "1000k", "-bufsize", "2000k", "-force_key_frames", "source", "-sc_threshold", "0", "-g", "300", "-r", "30",
		"-muxdelay", "0", "-muxpreload", "0", "-flush_packets", "1", "-f", "mpegts", "pipe:4",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("args mismatch\n got: %q\nwant: %q", got, want)
	}
}

func TestEncoderArgsSingle(t *testing.T) {
	got := EncoderArgs(landscape30(), []Rendition{encodedRendition(480, 30)}, "ultrafast", 28)

	want := []string{
		"-hide_banner", "-loglevel", "error", "-probesize", "32", "-analyzeduration", "0", "-f", "flv", "-i", "pipe:0", "-an", "-copyts",
		"-filter_complex", "[0:v]scale=-2:480[v0]",
		"-map", "[v0]", "-c:v", "libx264", "-preset", "ultrafast", "-tune", "zerolatency", "-pix_fmt", "yuv420p",
		"-crf", "28", "-maxrate", "1000k", "-bufsize", "2000k", "-force_key_frames", "source", "-sc_threshold", "0", "-g", "300", "-r", "30",
		"-muxdelay", "0", "-muxpreload", "0", "-flush_packets", "1", "-f", "mpegts", "pipe:3",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("args mismatch\n got: %q\nwant: %q", got, want)
	}
}

func TestEncoderArgsPortrait(t *testing.T) {
	got := EncoderArgs(SourceInfo{Width: 1080, Height: 1920, FPS: 30},
		[]Rendition{encodedRendition(720, 30), encodedRendition(480, 30)}, "veryfast", 23)

	graph := graphArg(t, got)
	want := "[0:v]split=2[s0][s1];[s0]scale=720:-2[v0];[s1]scale=480:-2[v1]"
	if graph != want {
		t.Errorf("graph = %q, want %q", graph, want)
	}
}

// The output rate is always set when the source rate is known: without it
// x264 takes its VUI timing from ffmpeg's guess on the metadata-less FLV
// input (60 fps for a 30 fps source), and the master advertises that rate.
func TestEncoderArgsFPS(t *testing.T) {
	same := EncoderArgs(landscape30(), []Rendition{encodedRendition(480, 30)}, "veryfast", 23)
	if g := indexOf(same, "-g"); g < 0 || g+3 >= len(same) || same[g+2] != "-r" || same[g+3] != "30" {
		t.Errorf("-r 30 must follow -g 300 for an equal fps: %q", same)
	}

	lower := EncoderArgs(landscape30(), []Rendition{encodedRendition(480, 24)}, "veryfast", 23)
	g := indexOf(lower, "-g")
	if g < 0 || g+3 >= len(lower) || lower[g+1] != "300" || lower[g+2] != "-r" || lower[g+3] != "24" {
		t.Errorf("-r 24 must follow -g 300: %q", lower)
	}

	unknown := EncoderArgs(SourceInfo{Width: 1280, Height: 720}, []Rendition{encodedRendition(480, 24)}, "veryfast", 23)
	if indexOf(unknown, "-r") >= 0 {
		t.Errorf("unexpected -r when source fps unknown: %q", unknown)
	}
}

func graphArg(t *testing.T, args []string) string {
	t.Helper()
	i := indexOf(args, "-filter_complex")
	if i < 0 || i+1 >= len(args) {
		t.Fatalf("no -filter_complex in %q", args)
	}
	return args[i+1]
}

func indexOf(args []string, s string) int {
	for i, a := range args {
		if a == s {
			return i
		}
	}
	return -1
}

func TestEncoderEndToEnd(t *testing.T) {
	requireFFmpeg(t)

	pkts := readFLV(t, "src_1280x720.flv")
	var firstKeyPTS90k int64 = -1
	var sourceFrames int
	for _, pkt := range pkts {
		if pkt.Type != av.H264 {
			continue
		}
		sourceFrames++
		if firstKeyPTS90k < 0 && pkt.IsKeyFrame {
			firstKeyPTS90k = (pkt.Time + pkt.CTime).Milliseconds() * 90
		}
	}
	if firstKeyPTS90k < 0 {
		t.Fatal("fixture has no keyframe")
	}

	src := SourceInfo{Width: 1280, Height: 720, FPS: 30}
	args := EncoderArgs(src, []Rendition{encodedRendition(480, 30)}, "veryfast", 23)

	ctx, cancel := context.WithTimeout(context.Background(), testEncoderTimeout)
	defer cancel()

	enc, err := StartEncoder(ctx, "test-stream", ffmpegBinary(t), args, 1)
	if err != nil {
		t.Fatalf("StartEncoder: %v", err)
	}
	t.Cleanup(enc.Stop)

	var ts bytes.Buffer
	readDone := make(chan error, 1)
	go func() {
		_, err := io.Copy(&ts, enc.Outputs()[0])
		readDone <- err
	}()

	mux := flv.NewMuxer(enc.Input())
	for _, pkt := range pkts {
		if err := mux.WritePacket(pkt); err != nil {
			t.Fatalf("write packet: %v", err)
		}
	}
	if err := enc.Input().Close(); err != nil {
		t.Fatalf("close input: %v", err)
	}

	if err := <-readDone; err != nil {
		t.Fatalf("read output: %v", err)
	}
	if err := enc.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}

	reader := &mpegts.Reader{R: &ts}
	if err := reader.Initialize(); err != nil {
		t.Fatalf("mpegts init: %v", err)
	}
	tracks := reader.Tracks()
	if len(tracks) != 1 {
		t.Fatalf("tracks = %d, want exactly 1 (video only)", len(tracks))
	}
	if _, ok := tracks[0].Codec.(*codecs.H264); !ok {
		t.Fatalf("track codec = %T, want *codecs.H264", tracks[0].Codec)
	}

	var (
		firstPTS  int64 = -1
		firstDTS  int64
		spsHeight int
		frames    int
	)
	reader.OnDataH264(tracks[0], func(pts, dts int64, au [][]byte) error {
		frames++
		if firstPTS < 0 {
			firstPTS, firstDTS = pts, dts
		}
		for _, nalu := range au {
			if spsHeight != 0 || len(nalu) == 0 || h264.NALUType(nalu[0]&0x1f) != h264.NALUTypeSPS {
				continue
			}
			var sps h264.SPS
			if err := sps.Unmarshal(nalu); err != nil {
				return err
			}
			spsHeight = sps.Height()
		}
		return nil
	})
	for {
		err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("mpegts read: %v", err)
		}
	}

	if frames == 0 {
		t.Fatal("no video frames in output")
	}
	if frames != sourceFrames {
		t.Errorf("output frames = %d, want %d (every source frame, none dropped at startup)", frames, sourceFrames)
	}
	if spsHeight != 480 {
		t.Errorf("SPS height = %d, want 480", spsHeight)
	}
	delta := firstPTS - firstKeyPTS90k
	t.Logf("source first keyframe PTS = %d, output first PTS = %d, DTS = %d, offset = %d ticks (%d frames)",
		firstKeyPTS90k, firstPTS, firstDTS, delta, frames)
	if delta < -3000 || delta > 3000 {
		t.Errorf("first PTS %d differs from source keyframe PTS %d by %d ticks, want within 3000", firstPTS, firstKeyPTS90k, delta)
	}
}

func ffmpegBinary(t *testing.T) string {
	t.Helper()
	p, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Fatalf("ffmpeg: %v", err)
	}
	return p
}

func TestStartEncoderBadPath(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-ffmpeg")
	enc, err := StartEncoder(context.Background(), "test-stream", missing, []string{"-version"}, 2)
	if err == nil {
		enc.Stop()
		t.Fatal("expected error for missing binary")
	}
	if enc != nil {
		t.Errorf("encoder = %v, want nil on error", enc)
	}
}

func TestEncoderStopKillsAndIsIdempotent(t *testing.T) {
	// A child that ignores nothing: sleep dies on SIGTERM.
	ctx, cancel := context.WithTimeout(context.Background(), testEncoderTimeout)
	defer cancel()

	enc, err := StartEncoder(ctx, "test-stream", "/bin/sh", []string{"-c", "sleep 30"}, 1)
	if err != nil {
		t.Fatalf("StartEncoder: %v", err)
	}

	start := time.Now()
	enc.Stop()
	enc.Stop()
	if err := enc.Wait(); err == nil {
		t.Error("Wait after Stop = nil, want signal error")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("Stop took %v, want prompt SIGTERM exit", elapsed)
	}
	enc.Stop() // after Wait

	if _, err := enc.Outputs()[0].Read(make([]byte, 1)); err == nil || errors.Is(err, io.EOF) {
		t.Errorf("read after Stop = %v, want closed-file error", err)
	}
}

func TestEncoderStopEscalatesToSIGKILL(t *testing.T) {
	script := filepath.Join(t.TempDir(), "stubborn.sh")
	ready := filepath.Join(t.TempDir(), "ready")
	body := "#!/bin/sh\ntrap '' TERM\ntouch \"$1\"\nwhile :; do sleep 1; done\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), testEncoderTimeout)
	defer cancel()

	enc, err := StartEncoder(ctx, "test-stream", script, []string{ready}, 0)
	if err != nil {
		t.Fatalf("StartEncoder: %v", err)
	}
	// SIGTERM sent before the trap is installed would end the script early.
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		if time.Now().After(deadline) {
			enc.Stop()
			t.Fatal("stub script never installed its SIGTERM trap")
		}
		time.Sleep(10 * time.Millisecond)
	}

	start := time.Now()
	enc.Stop()
	elapsed := time.Since(start)
	if elapsed < stopGracePeriod {
		t.Errorf("Stop returned after %v, before the %v grace period", elapsed, stopGracePeriod)
	}
	if err := enc.Wait(); err == nil {
		t.Error("Wait = nil, want kill error")
	}
}

// Every ffmpeg stderr line is attributed to its livestream, so concurrent
// streams can be told apart in the logs.
func TestEncoderStderrCarriesLivestreamID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testEncoderTimeout)
	defer cancel()

	const id = "stderr-id-0f3a"
	marker := "stderr-marker-" + t.Name()
	enc, err := StartEncoder(ctx, id, "/bin/sh", []string{"-c", "echo " + marker + "-1 >&2; echo " + marker + "-2 >&2"}, 0)
	if err != nil {
		t.Fatalf("StartEncoder: %v", err)
	}
	_ = enc.Wait()
	enc.Stop()

	lines := logsContaining(marker)
	if len(lines) != 2 {
		t.Fatalf("logged %d stderr line(s), want 2: %+v", len(lines), lines)
	}
	for _, l := range lines {
		if !strings.Contains(l.Message, "livestream "+id+": ffmpeg: ") {
			t.Errorf("stderr line %q lacks the livestream id", l.Message)
		}
	}
}
