package live

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"sen1or/letslive/transcode/config"
)

func TestEncoderArgsFileProfile(t *testing.T) {
	p := fileProfile(config.Upload{Threads: 2, Niceness: 10})
	got := EncoderArgs(landscape30(), []Rendition{encodedRendition(720, 30), encodedRendition(480, 30)}, "veryfast", 23, p)
	joined := strings.Join(got, " ")
	for _, banned := range []string{"zerolatency", "-probesize", "-analyzeduration", "-flush_packets"} {
		if strings.Contains(joined, banned) {
			t.Errorf("file profile args contain %q:\n%s", banned, joined)
		}
	}
	if n := strings.Count(joined, "-threads 2"); n != 2 {
		t.Errorf("-threads 2 appears %d time(s), want once per output:\n%s", n, joined)
	}
	if !strings.Contains(joined, "-preset veryfast") || !strings.Contains(joined, "-crf 23") {
		t.Errorf("preset/crf missing:\n%s", joined)
	}
}

func TestEncoderArgsLiveProfileHasNoThreads(t *testing.T) {
	got := EncoderArgs(landscape30(), []Rendition{encodedRendition(480, 30)}, "veryfast", 23, liveProfile)
	if slices.Contains(got, "-threads") {
		t.Errorf("live args set -threads: %v", got)
	}
}

// Review Focus 4: an upload's ffmpeg must not compete with live encoders at
// the same priority.
func TestSetNicenessApplies(t *testing.T) {
	cmd := exec.Command("sleep", "5")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })

	if err := SetNiceness(cmd.Process.Pid, 5); err != nil {
		t.Fatalf("SetNiceness: %v", err)
	}
	if got := niceOf(t, cmd.Process.Pid); got != 5 {
		t.Errorf("nice = %d, want 5", got)
	}
}

func TestSetNicenessZeroIsNoop(t *testing.T) {
	if err := SetNiceness(-1, 0); err != nil {
		t.Errorf("SetNiceness(_, 0) = %v, want nil", err)
	}
}

// niceOf reads a process's nice value; Linux's raw getpriority returns
// 20-nice, macOS returns nice itself.
func niceOf(t *testing.T, pid int) int {
	t.Helper()
	prio, err := syscall.Getpriority(syscall.PRIO_PROCESS, pid)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "linux" {
		return 20 - prio
	}
	return prio
}

// Review Focus 2: draining keeps the last GOP of every encoded rendition.
func TestDrainKeepsEncodedTail(t *testing.T) {
	requireFFmpeg(t)
	rig := newSessionRig(t, ffmpegBinary(t))
	s, fail := newFileishSession(t, rig)

	feedSession(t, s, readFLV(t, "src_1280x720.flv"))
	if err := closeSession(t, s); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := fail.Err(); err != nil {
		t.Fatalf("failure reported: %v", err)
	}

	vod := filepath.Join(rig.tc.PublicHLSPath, "vods", s.id)
	src := playlistSeconds(t, filepath.Join(vod, "0", "stream.m3u8"))
	enc := playlistSeconds(t, filepath.Join(vod, "1", "stream.m3u8"))
	tolerance := float64(rig.tc.Live.RecordSegmentDurationMs) / 1000
	if src-enc > tolerance || enc-src > tolerance {
		t.Errorf("encoded rendition lasts %.2fs, source %.2fs; want within %.1fs", enc, src, tolerance)
	}
	t.Logf("source %.2fs, encoded %.2fs", src, enc)
}

func playlistSeconds(t *testing.T, path string) float64 {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var total float64
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line, ok := strings.CutPrefix(sc.Text(), "#EXTINF:")
		if !ok {
			continue
		}
		v, err := strconv.ParseFloat(strings.TrimSuffix(line, ","), 64)
		if err != nil {
			t.Fatalf("EXTINF %q: %v", line, err)
		}
		total += v
	}
	return total
}
