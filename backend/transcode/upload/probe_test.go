package upload

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestParseStreams(t *testing.T) {
	cases := []struct {
		file string
		want Probe
	}{
		{"h264.streams.json", Probe{HasVideo: true, VideoCodec: "h264", PixFmt: "yuv420p", FieldOrder: "progressive",
			Width: 1280, Height: 720, FrameRate: 30, HasAudio: true, AudioCodec: "aac", AudioChannels: 1, Duration: 3 * time.Second}},
		{"rot.streams.json", Probe{HasVideo: true, VideoCodec: "h264", PixFmt: "yuv420p", FieldOrder: "progressive",
			Width: 1280, Height: 720, Rotation: 90, FrameRate: 30, HasAudio: true, AudioCodec: "aac", AudioChannels: 1, Duration: 3 * time.Second}},
		{"hlg.streams.json", Probe{HasVideo: true, VideoCodec: "hevc", PixFmt: "yuv420p10le", FieldOrder: "progressive",
			ColorTransfer: "arib-std-b67", Width: 1920, Height: 1080, FrameRate: 30, HasAudio: true, AudioCodec: "aac", AudioChannels: 1}},
		{"interlaced.streams.json", Probe{HasVideo: true, VideoCodec: "h264", PixFmt: "yuv420p", FieldOrder: "tt",
			Width: 720, Height: 576, FrameRate: 25}},
		{"vp9.streams.json", Probe{HasVideo: true, VideoCodec: "vp9", PixFmt: "yuv420p", FieldOrder: "progressive",
			Width: 640, Height: 360, FrameRate: 30, HasAudio: true, AudioCodec: "opus", AudioChannels: 1}},
		{"six.streams.json", Probe{HasVideo: true, VideoCodec: "h264", PixFmt: "yuv420p", FieldOrder: "progressive",
			Width: 640, Height: 360, FrameRate: 30, HasAudio: true, AudioCodec: "aac", AudioChannels: 6}},
	}
	for _, c := range cases {
		t.Run(c.file, func(t *testing.T) {
			got, err := parseStreams(readTestdata(t, c.file))
			if err != nil {
				t.Fatal(err)
			}
			// Container durations vary by a few ms between muxers; only
			// check them where the case pins one.
			if c.want.Duration == 0 {
				got.Duration = 0
			} else if d := got.Duration - c.want.Duration; d < -100*time.Millisecond || d > 100*time.Millisecond {
				t.Errorf("Duration = %v, want ~%v", got.Duration, c.want.Duration)
			}
			got.Duration, c.want.Duration = 0, 0
			if got != c.want {
				t.Errorf("parseStreams =\n %+v\nwant\n %+v", got, c.want)
			}
		})
	}
}

func TestParseStreamsNoVideo(t *testing.T) {
	got, err := parseStreams([]byte(`{"streams":[{"codec_type":"audio","codec_name":"mp3","channels":2}],"format":{"duration":"4.0"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if got.HasVideo || !got.HasAudio || got.AudioCodec != "mp3" {
		t.Errorf("parseStreams = %+v, want audio only", got)
	}
}

func TestParseStreamsRotateTag(t *testing.T) {
	got, err := parseStreams([]byte(`{"streams":[{"codec_type":"video","codec_name":"h264","width":1920,"height":1080,"tags":{"rotate":"-90"}}],"format":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	if got.Rotation != -90 {
		t.Errorf("Rotation = %d, want -90", got.Rotation)
	}
}

// Review finding: a VFR file's average rate (17.3 here: frames dropped in
// gaps) is not its motion rate; converting at it would look choppy.
func TestParseStreamsVFRUsesBaseRate(t *testing.T) {
	got, err := parseStreams(readTestdata(t, "vfr.streams.json"))
	if err != nil {
		t.Fatal(err)
	}
	if got.FrameRate != 30 {
		t.Errorf("FrameRate = %v, want 30 (r_frame_rate; avg_frame_rate is 294/17)", got.FrameRate)
	}
}

func TestMaxKeyframeGap(t *testing.T) {
	gap, err := maxKeyframeGap(readTestdata(t, "h264.keyframes.json"), 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if gap != time.Second {
		t.Errorf("gap = %v, want 1s", gap)
	}

	// The tail after the last keyframe counts as a gap.
	gap, err = maxKeyframeGap([]byte(`{"frames":[{"pts_time":"0.0"},{"pts_time":"1.0"}]}`), 11*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if gap != 10*time.Second {
		t.Errorf("gap with a long tail = %v, want 10s", gap)
	}

	if _, err := maxKeyframeGap([]byte(`{"frames":[]}`), time.Second); err == nil {
		t.Error("no keyframe = nil error, want one")
	}
}

func TestFFprobePath(t *testing.T) {
	for in, want := range map[string]string{
		"ffmpeg":                   "ffprobe",
		"/usr/bin/ffmpeg":          "/usr/bin/ffprobe",
		"/opt/homebrew/bin/ffmpeg": "/opt/homebrew/bin/ffprobe",
	} {
		if got := ffprobePath(in); got != want {
			t.Errorf("ffprobePath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRunProbe(t *testing.T) {
	ffmpeg := requireFFmpeg(t)
	in := filepath.Join(t.TempDir(), "in.mp4")
	genFile(t, ffmpeg, in, "-f", "lavfi", "-i", "testsrc2=s=640x360:r=30:d=4",
		"-f", "lavfi", "-i", "sine=d=4", "-c:v", "libx264", "-g", "60", "-c:a", "aac", "-shortest")

	p, err := RunProbe(t.Context(), ffmpeg, in)
	if err != nil {
		t.Fatal(err)
	}
	if !p.HasVideo || p.VideoCodec != "h264" || p.Width != 640 || !p.HasAudio || p.AudioCodec != "aac" {
		t.Errorf("RunProbe = %+v", p)
	}
	if p.MaxKeyframeGap < 1900*time.Millisecond || p.MaxKeyframeGap > 2100*time.Millisecond {
		t.Errorf("MaxKeyframeGap = %v, want ~2s (GOP 60 at 30 fps)", p.MaxKeyframeGap)
	}
}
