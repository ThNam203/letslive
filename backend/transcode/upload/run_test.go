package upload

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"

	"sen1or/letslive/transcode/config"
)

const runTestSeconds = 10

// dirStorage stores uploads under dir by their object path, like MinIO would.
type dirStorage struct {
	dir string

	mu     sync.Mutex
	thumbs []string
}

func (s *dirStorage) AddSegment(_ context.Context, path, id string, q int) (string, error) {
	return s.put(path, filepath.Join(id, fmt.Sprint(q), filepath.Base(path)))
}

func (s *dirStorage) AddThumbnail(_ context.Context, path, id, _ string) (string, error) {
	url, err := s.put(path, filepath.Join(id, filepath.Base(path)))
	if err == nil {
		s.mu.Lock()
		s.thumbs = append(s.thumbs, url)
		s.mu.Unlock()
	}
	return url, err
}

func (s *dirStorage) put(src, object string) (string, error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	dst := filepath.Join(s.dir, object)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", err
	}
	return "http://storage.test/" + filepath.ToSlash(object), os.WriteFile(dst, data, 0o644)
}

type runRig struct {
	ffmpeg string
	tc     config.Transcode
	st     *dirStorage
}

func newRunRig(t *testing.T) *runRig {
	t.Helper()
	ffmpeg := requireFFmpeg(t)
	var tc config.Transcode
	tc.PublicHLSPath = t.TempDir()
	tc.PrivateHLSPath = t.TempDir()
	tc.FFMpegSetting.FFMpegPath = ffmpeg
	tc.FFMpegSetting.MasterFileName = "index.m3u8"
	tc.FFMpegSetting.CRF = 28
	tc.FFMpegSetting.Preset = "ultrafast"
	tc.FFMpegSetting.Qualities = []config.Quality{
		{Resolution: "1920x1080", MaxBitrate: "5000k", FPS: 30, BufSize: "10000k"},
		{Resolution: "1280x720", MaxBitrate: "2500k", FPS: 30, BufSize: "5000k"},
		{Resolution: "640x360", MaxBitrate: "800k", FPS: 30, BufSize: "1600k"},
	}
	tc.Live = config.Live{}.WithDefaults()
	tc.Upload = config.Upload{}.WithDefaults()
	return &runRig{ffmpeg: ffmpeg, tc: tc, st: &dirStorage{dir: t.TempDir()}}
}

// input writes a runTestSeconds test file with the given ffmpeg args.
func (r *runRig) input(t *testing.T, name string, args ...string) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), name)
	genFile(t, r.ffmpeg, out, args...)
	return out
}

func lavfiVideo(size string) []string {
	return []string{"-f", "lavfi", "-i", fmt.Sprintf("testsrc2=s=%s:r=30:d=%d", size, runTestSeconds)}
}

func lavfiAudio() []string {
	return []string{"-f", "lavfi", "-i", fmt.Sprintf("sine=frequency=440:d=%d", runTestSeconds)}
}

func join(parts ...[]string) []string {
	var out []string
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

type segProbe struct {
	VideoCodec, PixFmt, Transfer string
	Width, Height                int
	HasAudio                     bool
	AudioCodec                   string
}

// probeSegment ffprobes the first stored segment of rendition q.
func (r *runRig) probeSegment(t *testing.T, id string, q int) segProbe {
	t.Helper()
	seg := filepath.Join(r.st.dir, id, fmt.Sprint(q), "0.ts")
	out, err := exec.Command(ffprobePath(r.ffmpeg), "-v", "error", "-of", "json", "-show_streams", seg).Output()
	if err != nil {
		t.Fatalf("ffprobe %s: %v", seg, err)
	}
	var j probeJSON
	if err := json.Unmarshal(out, &j); err != nil {
		t.Fatal(err)
	}
	var p segProbe
	for _, s := range j.Streams {
		switch s.CodecType {
		case "video":
			p.VideoCodec, p.PixFmt, p.Transfer, p.Width, p.Height = s.CodecName, s.PixFmt, s.ColorTransfer, s.Width, s.Height
		case "audio":
			p.HasAudio, p.AudioCodec = true, s.CodecName
		}
	}
	return p
}

// run runs Run and checks what every successful upload must have.
func (r *runRig) run(t *testing.T, input string) (string, Result) {
	t.Helper()
	id := uuid.Must(uuid.NewV4()).String()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	res, err := Run(ctx, r.tc, r.st, id, input, false)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	master, err := os.ReadFile(filepath.Join(r.tc.PublicHLSPath, "vods", id, "index.m3u8"))
	if err != nil {
		t.Fatalf("VOD master: %v", err)
	}
	if !strings.Contains(string(master), "CODECS=") || !strings.Contains(string(master), "0/stream.m3u8") {
		t.Errorf("master lacks CODECS or rendition 0:\n%s", master)
	}
	if d := res.Duration.Seconds(); d < runTestSeconds-1 || d > runTestSeconds+1 {
		t.Errorf("Duration = %v, want %ds ± 1", res.Duration, runTestSeconds)
	}
	if res.ThumbnailURL == "" {
		t.Error("ThumbnailURL is empty")
	}
	return id, res
}

func variants(t *testing.T, r *runRig, id string) int {
	t.Helper()
	master, err := os.ReadFile(filepath.Join(r.tc.PublicHLSPath, "vods", id, "index.m3u8"))
	if err != nil {
		t.Fatal(err)
	}
	return strings.Count(string(master), "#EXT-X-STREAM-INF")
}

func TestRunH264AACIsCopied(t *testing.T) {
	r := newRunRig(t)
	in := r.input(t, "in.mp4", join(lavfiVideo("1280x720"), lavfiAudio(),
		[]string{"-c:v", "libx264", "-g", "60", "-c:a", "aac", "-shortest"})...)
	id, res := r.run(t, in)
	if !res.Plan.CopyVideo || !res.Plan.CopyAudio {
		t.Errorf("plan = %+v, want video and audio copied", res.Plan)
	}
	if n := variants(t, r, id); n != 2 {
		t.Errorf("variants = %d, want 2 (720p source + 360p)", n)
	}
	seg := r.probeSegment(t, id, 1)
	if seg.Height != 360 || !seg.HasAudio {
		t.Errorf("rendition 1 = %+v, want 360p with audio", seg)
	}
}

func TestRunHEVCIsConverted(t *testing.T) {
	r := newRunRig(t)
	in := r.input(t, "in.mp4", join(lavfiVideo("1280x720"), lavfiAudio(),
		[]string{"-c:v", "libx265", "-x265-params", "log-level=error", "-c:a", "aac", "-shortest"})...)
	id, res := r.run(t, in)
	if res.Plan.CopyVideo {
		t.Error("HEVC video was copied")
	}
	if seg := r.probeSegment(t, id, 0); seg.VideoCodec != "h264" || seg.Width != 1280 {
		t.Errorf("rendition 0 = %+v, want 1280-wide h264", seg)
	}
}

func TestRunVP9OpusWebM(t *testing.T) {
	r := newRunRig(t)
	in := r.input(t, "in.webm", join(lavfiVideo("640x360"), lavfiAudio(),
		[]string{"-c:v", "libvpx-vp9", "-deadline", "realtime", "-c:a", "libopus", "-shortest"})...)
	id, _ := r.run(t, in)
	if seg := r.probeSegment(t, id, 0); seg.VideoCodec != "h264" || seg.AudioCodec != "aac" {
		t.Errorf("rendition 0 = %+v, want h264 + aac", seg)
	}
}

// Review Focus 3: a phone's portrait video, stored landscape with a rotation.
func TestRunRotatedPortraitComesOutUpright(t *testing.T) {
	r := newRunRig(t)
	plain := r.input(t, "plain.mp4", join(lavfiVideo("1280x720"), lavfiAudio(),
		[]string{"-c:v", "libx264", "-c:a", "aac", "-shortest"})...)
	in := filepath.Join(t.TempDir(), "rot.mp4")
	genFile(t, r.ffmpeg, in, "-display_rotation", "90", "-i", plain, "-c", "copy")

	id, res := r.run(t, in)
	if res.Plan.CopyVideo {
		t.Error("rotated video was copied")
	}
	if seg := r.probeSegment(t, id, 0); seg.Width != 720 || seg.Height != 1280 {
		t.Errorf("rendition 0 = %dx%d, want 720x1280", seg.Width, seg.Height)
	}
}

func TestRunLargeIsScaledTo1080(t *testing.T) {
	for _, c := range []struct{ size, want string }{{"2560x1440", "1920x1080"}, {"1440x2560", "1080x1920"}} {
		t.Run(c.size, func(t *testing.T) {
			r := newRunRig(t)
			in := r.input(t, "in.mp4", join(lavfiVideo(c.size), lavfiAudio(),
				[]string{"-c:v", "libx264", "-preset", "ultrafast", "-c:a", "aac", "-shortest"})...)
			id, _ := r.run(t, in)
			seg := r.probeSegment(t, id, 0)
			if got := fmt.Sprintf("%dx%d", seg.Width, seg.Height); got != c.want {
				t.Errorf("rendition 0 = %s, want %s", got, c.want)
			}
		})
	}
}

func TestRunWithoutAudio(t *testing.T) {
	r := newRunRig(t)
	in := r.input(t, "in.mp4", join(lavfiVideo("1280x720"), []string{"-c:v", "libx264", "-g", "60"})...)
	id, res := r.run(t, in)
	if res.Plan.HasAudio {
		t.Error("plan has audio")
	}
	if seg := r.probeSegment(t, id, 0); seg.HasAudio {
		t.Error("segment has an audio stream")
	}
}

func TestRunWithoutVideoFails(t *testing.T) {
	r := newRunRig(t)
	in := r.input(t, "in.m4a", join(lavfiAudio(), []string{"-c:a", "aac"})...)
	_, err := Run(context.Background(), r.tc, r.st, uuid.Must(uuid.NewV4()).String(), in, false)
	if !errors.Is(err, ErrInput) || !errors.Is(err, ErrNoVideo) {
		t.Errorf("Run = %v, want ErrInput and ErrNoVideo", err)
	}
}

func TestRunUnreadableFileFails(t *testing.T) {
	r := newRunRig(t)
	in := filepath.Join(t.TempDir(), "junk.mp4")
	if err := os.WriteFile(in, []byte("not a video"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Run(context.Background(), r.tc, r.st, uuid.Must(uuid.NewV4()).String(), in, false)
	if !errors.Is(err, ErrInput) {
		t.Errorf("Run = %v, want ErrInput", err)
	}
}

// Review Focus 3: iPhone HDR (HLG, 10-bit) must come out as 8-bit SDR BT.709.
func TestRunHDRIsToneMapped(t *testing.T) {
	r := newRunRig(t)
	out, err := exec.Command(r.ffmpeg, "-hide_banner", "-filters").Output()
	if err != nil || !strings.Contains(string(out), " zscale ") {
		t.Skip("this ffmpeg has no zscale filter; the production image's has")
	}
	in := r.input(t, "hdr.mp4", join(lavfiVideo("1920x1080"), lavfiAudio(), []string{
		"-c:v", "libx265", "-pix_fmt", "yuv420p10le",
		"-x265-params", "log-level=error:colorprim=bt2020:transfer=arib-std-b67:colormatrix=bt2020nc",
		"-c:a", "aac", "-shortest"})...)
	id, res := r.run(t, in)
	if !res.Plan.ToneMap {
		t.Errorf("plan = %+v, want ToneMap", res.Plan)
	}
	if seg := r.probeSegment(t, id, 0); seg.PixFmt != "yuv420p" || seg.Transfer != "bt709" {
		t.Errorf("rendition 0 = %+v, want yuv420p bt709", seg)
	}
}
