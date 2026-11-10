package upload

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"sen1or/letslive/transcode/config"
)

const testMaxGap = 8 * time.Second

var testQualities = []config.Quality{
	{Resolution: "640x360", MaxBitrate: "365k", FPS: 30, BufSize: "730k"},
	{Resolution: "1280x720", MaxBitrate: "2.0M", FPS: 30, BufSize: "4M"},
	{Resolution: "1920x1080", MaxBitrate: "3.0M", FPS: 30, BufSize: "6.0M"},
}

func goodH264() Probe {
	return Probe{HasVideo: true, VideoCodec: "h264", PixFmt: "yuv420p", FieldOrder: "progressive",
		Width: 1920, Height: 1080, FrameRate: 30, MaxKeyframeGap: 2 * time.Second,
		HasAudio: true, AudioCodec: "aac", AudioChannels: 2}
}

func TestDecide(t *testing.T) {
	with := func(f func(*Probe)) Probe { p := goodH264(); f(&p); return p }
	cases := []struct {
		name  string
		probe Probe
		force bool
		want  Plan
	}{
		{"h264 1080p copies", goodH264(), false,
			Plan{CopyVideo: true, CopyAudio: true, HasAudio: true, FPS: 30}},
		{"portrait 1080x1920 copies", with(func(p *Probe) { p.Width, p.Height = 1080, 1920 }), false,
			Plan{CopyVideo: true, CopyAudio: true, HasAudio: true, FPS: 30}},
		{"yuvj420p copies", with(func(p *Probe) { p.PixFmt = "yuvj420p" }), false,
			Plan{CopyVideo: true, CopyAudio: true, HasAudio: true, FPS: 30}},
		{"hevc converts", with(func(p *Probe) { p.VideoCodec = "hevc" }), false,
			Plan{CopyAudio: true, HasAudio: true, FPS: 30, MaxRate: "3.0M", BufSize: "6.0M"}},
		{"10-bit converts", with(func(p *Probe) { p.PixFmt = "yuv420p10le" }), false,
			Plan{CopyAudio: true, HasAudio: true, FPS: 30, MaxRate: "3.0M", BufSize: "6.0M"}},
		{"HLG tone-maps", with(func(p *Probe) { p.VideoCodec, p.PixFmt, p.ColorTransfer = "hevc", "yuv420p10le", "arib-std-b67" }), false,
			Plan{CopyAudio: true, HasAudio: true, ToneMap: true, FPS: 30, MaxRate: "3.0M", BufSize: "6.0M"}},
		{"PQ tone-maps", with(func(p *Probe) { p.ColorTransfer = "smpte2084" }), false,
			Plan{CopyAudio: true, HasAudio: true, ToneMap: true, FPS: 30, MaxRate: "3.0M", BufSize: "6.0M"}},
		{"rotated converts", with(func(p *Probe) { p.Rotation = 90 }), false,
			Plan{CopyAudio: true, HasAudio: true, FPS: 30, MaxRate: "3.0M", BufSize: "6.0M"}},
		{"rotated 360 copies", with(func(p *Probe) { p.Rotation = -360 }), false,
			Plan{CopyVideo: true, CopyAudio: true, HasAudio: true, FPS: 30}},
		{"interlaced deinterlaces", with(func(p *Probe) { p.Width, p.Height, p.FieldOrder = 720, 576, "tt" }), false,
			Plan{CopyAudio: true, HasAudio: true, Deinterlace: true, FPS: 30, MaxRate: "2.0M", BufSize: "4M"}},
		{"1440p scales to 1080", with(func(p *Probe) { p.Width, p.Height = 2560, 1440 }), false,
			Plan{CopyAudio: true, HasAudio: true, ScaleShort: 1080, FPS: 30, MaxRate: "3.0M", BufSize: "6.0M"}},
		{"portrait 4K scales to 1080", with(func(p *Probe) { p.Width, p.Height = 2160, 3840 }), false,
			Plan{CopyAudio: true, HasAudio: true, ScaleShort: 1080, FPS: 30, MaxRate: "3.0M", BufSize: "6.0M"}},
		{"9s keyframe gap converts", with(func(p *Probe) { p.MaxKeyframeGap = 9 * time.Second }), false,
			Plan{CopyAudio: true, HasAudio: true, FPS: 30, MaxRate: "3.0M", BufSize: "6.0M"}},
		{"8s keyframe gap copies", with(func(p *Probe) { p.MaxKeyframeGap = testMaxGap }), false,
			Plan{CopyVideo: true, CopyAudio: true, HasAudio: true, FPS: 30}},
		{"forced converts", goodH264(), true,
			Plan{HasAudio: true, FPS: 30, MaxRate: "3.0M", BufSize: "6.0M"}},
		{"opus converts audio", with(func(p *Probe) { p.AudioCodec = "opus" }), false,
			Plan{CopyVideo: true, HasAudio: true, FPS: 30}},
		{"5.1 aac converts to stereo", with(func(p *Probe) { p.AudioChannels = 6 }), false,
			Plan{CopyVideo: true, HasAudio: true, Stereo: true, FPS: 30}},
		{"5.1 opus converts to stereo", with(func(p *Probe) { p.AudioCodec, p.AudioChannels = "opus", 6 }), false,
			Plan{CopyVideo: true, HasAudio: true, Stereo: true, FPS: 30}},
		{"no audio", with(func(p *Probe) { p.HasAudio, p.AudioCodec, p.AudioChannels = false, "", 0 }), false,
			Plan{CopyVideo: true, FPS: 30}},
		{"29.97 fps rounds to 30", with(func(p *Probe) { p.VideoCodec, p.FrameRate = "hevc", 29.97 }), false,
			Plan{CopyAudio: true, HasAudio: true, FPS: 30, MaxRate: "3.0M", BufSize: "6.0M"}},
		{"120 fps caps at 60", with(func(p *Probe) { p.VideoCodec, p.FrameRate = "hevc", 120 }), false,
			Plan{CopyAudio: true, HasAudio: true, FPS: 60, MaxRate: "3.0M", BufSize: "6.0M"}},
		{"unknown fps is 30", with(func(p *Probe) { p.VideoCodec, p.FrameRate = "hevc", 0 }), false,
			Plan{CopyAudio: true, HasAudio: true, FPS: 30, MaxRate: "3.0M", BufSize: "6.0M"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Decide(c.probe, testMaxGap, c.force, testQualities)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("Decide =\n %+v\nwant\n %+v", got, c.want)
			}
		})
	}
}

func TestDecideNoVideo(t *testing.T) {
	_, err := Decide(Probe{HasAudio: true, AudioCodec: "aac"}, testMaxGap, false, testQualities)
	if !errors.Is(err, ErrNoVideo) {
		t.Errorf("Decide without video = %v, want ErrNoVideo", err)
	}
}

func TestIntakeArgsCopy(t *testing.T) {
	got := IntakeArgs("in.mp4", Plan{CopyVideo: true, CopyAudio: true, HasAudio: true, FPS: 30}, "veryfast", 23, 0)
	want := []string{
		"-hide_banner", "-loglevel", "error", "-i", "in.mp4",
		"-map", "0:v:0", "-map", "0:a:0",
		"-c:v", "copy",
		"-c:a", "copy",
		"-avoid_negative_ts", "make_zero", "-flvflags", "no_duration_filesize", "-f", "flv", "pipe:1",
	}
	if !slices.Equal(got, want) {
		t.Errorf("IntakeArgs =\n %q\nwant\n %q", got, want)
	}
}

func TestIntakeArgsFullConvert(t *testing.T) {
	pl := Plan{HasAudio: true, Stereo: true, ToneMap: true, Deinterlace: true, ScaleShort: 1080, FPS: 25, MaxRate: "3.0M", BufSize: "6.0M"}
	got := IntakeArgs("in.mov", pl, "veryfast", 23, 2)
	want := []string{
		"-hide_banner", "-loglevel", "error", "-i", "in.mov",
		"-map", "0:v:0", "-map", "0:a:0",
		"-vf", "yadif," + toneMapFilter + "," + scaleShortFilter(1080) + ",format=yuv420p",
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "23", "-maxrate", "3.0M", "-bufsize", "6.0M",
		"-fps_mode", "cfr", "-r", "25", "-force_key_frames", "expr:gte(t,n_forced*2)",
		"-c:a", "aac", "-b:a", "128k", "-ac", "2",
		"-threads", "2",
		"-avoid_negative_ts", "make_zero", "-flvflags", "no_duration_filesize", "-f", "flv", "pipe:1",
	}
	if !slices.Equal(got, want) {
		t.Errorf("IntakeArgs =\n %q\nwant\n %q", got, want)
	}
	if !strings.Contains(toneMapFilter, "tonemap=") || !strings.Contains(toneMapFilter, "zscale=") {
		t.Errorf("toneMapFilter = %q, want zscale + tonemap", toneMapFilter)
	}
}

func TestIntakeArgsVideoOnlyConvert(t *testing.T) {
	got := IntakeArgs("in.webm", Plan{FPS: 30}, "veryfast", 23, 0)
	if slices.Contains(got, "0:a:0") || slices.Contains(got, "-c:a") {
		t.Errorf("video-only plan maps audio: %q", got)
	}
	if i := slices.Index(got, "-vf"); i < 0 || got[i+1] != "format=yuv420p" {
		t.Errorf("-vf = %q, want just format=yuv420p", got)
	}
}

func TestRateCap(t *testing.T) {
	for _, c := range []struct {
		short          int
		maxRate, bufSz string
	}{{360, "365k", "730k"}, {576, "2.0M", "4M"}, {720, "2.0M", "4M"}, {1080, "3.0M", "6.0M"}, {1440, "3.0M", "6.0M"}} {
		if m, b := rateCap(testQualities, c.short); m != c.maxRate || b != c.bufSz {
			t.Errorf("rateCap(%d) = %s/%s, want %s/%s", c.short, m, b, c.maxRate, c.bufSz)
		}
	}
	if m, b := rateCap(nil, 720); m != "" || b != "" {
		t.Errorf("rateCap without qualities = %q/%q, want none", m, b)
	}
}
