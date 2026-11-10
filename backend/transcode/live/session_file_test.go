package live

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"

	"sen1or/letslive/transcode/config"
)

// newFileSession starts a Session in FileMode.
func newFileSession(t *testing.T, rig *sessionRig) *Session {
	t.Helper()
	m, cancel := FileMode(context.Background(), config.Upload{}.WithDefaults())
	t.Cleanup(cancel)
	s := NewSession(context.Background(), uuid.Must(uuid.NewV4()).String(), SessionDeps{
		Mode:      m,
		Storage:   rig.st,
		Transcode: rig.tc,
		Now:       time.Now,
	})
	t.Cleanup(func() { closeSession(t, s) })
	return s
}

// masterShape is the VOD master without what depends on the encoder's
// bitrate: one "RESOLUTION CODECS URI" line per variant, in order.
func masterShape(t *testing.T, path string) []string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var shape []string
	lines := strings.Split(string(body), "\n")
	for i, line := range lines {
		attrs, ok := strings.CutPrefix(line, "#EXT-X-STREAM-INF:")
		if !ok || i+1 >= len(lines) {
			continue
		}
		var keep []string
		for _, a := range strings.Split(attrs, ",") {
			if strings.HasPrefix(a, "RESOLUTION=") || strings.HasPrefix(a, "CODECS=") || strings.HasPrefix(a, `"`) || strings.HasSuffix(a, `"`) {
				keep = append(keep, a)
			}
		}
		shape = append(shape, strings.Join(keep, ",")+" "+lines[i+1])
	}
	return shape
}

func TestFileSessionWritesVODLikeLive(t *testing.T) {
	requireFFmpeg(t)
	rig := newSessionRig(t, ffmpegBinary(t))
	pkts := readFLV(t, "src_1280x720.flv")

	live, base := rig.newSession(t)
	feedSession(t, live, pkts)
	// Let the live encoder catch up, or Close cuts its rendition short.
	waitMaster(t, base, []string{"source", "480p"}, nil)
	if err := closeSession(t, live); err != nil {
		t.Fatalf("live Close: %v", err)
	}
	file := newFileSession(t, rig)
	feedSession(t, file, pkts)
	if err := closeSession(t, file); err != nil {
		t.Fatalf("file Close: %v", err)
	}

	vods := filepath.Join(rig.tc.PublicHLSPath, "vods")
	want := masterShape(t, filepath.Join(vods, live.id, sessionMasterFile))
	got := masterShape(t, filepath.Join(vods, file.id, sessionMasterFile))
	if len(want) < 2 || !slices.Equal(got, want) {
		t.Errorf("file VOD master = %q, want the live one %q", got, want)
	}
	if !strings.Contains(strings.Join(got, " "), "CODECS=") {
		t.Errorf("master lacks CODECS: %q", got)
	}
}

func TestFileSessionNotRegistered(t *testing.T) {
	rig := newSessionRig(t, filepath.Join(t.TempDir(), "no-ffmpeg"))
	s := newFileSession(t, rig)
	pkts := readFLV(t, "src_640x360.flv")
	feedSession(t, s, pkts)
	if registered(rig.reg, s.id) {
		t.Error("a file session was registered for live viewing")
	}
	if _, ok := s.out.(noOutput); !ok {
		t.Errorf("output = %T, want noOutput", s.out)
	}
}

func TestFileSessionCloseFailsOnEncoderFailure(t *testing.T) {
	rig := newSessionRig(t, filepath.Join(t.TempDir(), "no-ffmpeg"))
	s := newFileSession(t, rig)
	for _, pkt := range readFLV(t, "src_1280x720.flv") {
		if err := s.HandlePacket(pkt); err != nil {
			break
		}
	}
	err := closeSession(t, s)
	if err == nil || !strings.Contains(err.Error(), "encoder") {
		t.Errorf("Close = %v, want the encoder failure", err)
	}
}

func TestFileSessionReturnsThumbnailURL(t *testing.T) {
	requireFFmpeg(t)
	rig := newSessionRig(t, ffmpegBinary(t))
	s := newFileSession(t, rig)
	feedSession(t, s, readFLV(t, "src_640x360.flv"))
	if err := closeSession(t, s); err != nil {
		t.Fatalf("Close: %v", err)
	}
	var thumbs []string
	for _, c := range rig.st.Calls() {
		if c.Kind == "thumbnail" && c.StreamID == s.id {
			thumbs = append(thumbs, c.URL)
		}
	}
	if len(thumbs) != 1 || s.ThumbnailURL() != thumbs[0] {
		t.Errorf("ThumbnailURL = %q, uploaded %q; want exactly one, returned", s.ThumbnailURL(), thumbs)
	}
}

func TestFileSessionLogsUploadLabel(t *testing.T) {
	rig := newSessionRig(t, filepath.Join(t.TempDir(), "no-ffmpeg"))
	s := newFileSession(t, rig)
	feedSession(t, s, readFLV(t, "src_640x360.flv"))
	closeSession(t, s)
	for _, e := range logsContaining(s.id) {
		if !strings.HasPrefix(e.Message, "upload "+s.id) {
			t.Errorf("log line %q does not start with %q", e.Message, "upload "+s.id)
		}
	}
	if len(logsContaining("upload "+s.id)) == 0 {
		t.Error("no log line carries the upload label")
	}
}
