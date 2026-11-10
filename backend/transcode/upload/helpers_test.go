package upload

import (
	"os"
	"os/exec"
	"testing"

	"sen1or/letslive/shared/pkg/logger"
)

func TestMain(m *testing.M) {
	logger.Init(logger.Error)
	os.Exit(m.Run())
}

// requireFFmpeg returns the ffmpeg path, skipping the test when ffmpeg or
// ffprobe is missing.
func requireFFmpeg(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not installed")
	}
	if _, err := exec.LookPath(ffprobePath(path)); err != nil {
		t.Skip("ffprobe not installed next to ffmpeg")
	}
	return path
}

// genFile runs ffmpeg with args to write out.
func genFile(t *testing.T, ffmpeg, out string, args ...string) {
	t.Helper()
	cmd := exec.Command(ffmpeg, append(append([]string{"-hide_banner", "-loglevel", "error", "-y"}, args...), out)...)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generate %s: %v\n%s", out, err, b)
	}
}
