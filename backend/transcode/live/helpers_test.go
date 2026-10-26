package live

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sync"
	"testing"

	"github.com/nareix/joy5/av"
	"github.com/nareix/joy5/format/flv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"sen1or/letslive/shared/pkg/logger"
)

const recordedURLPrefix = "http://minio.test/bucket"

// testLogs records every line logged by the package's tests, at every level.
// The shared logger is set up once here, before any test runs, so no test
// races a later reassignment; lines at Error and above still reach stdout.
var testLogs *observer.ObservedLogs

func init() {
	logger.Init(logger.Error)
	core, logs := observer.New(zapcore.DebugLevel)
	testLogs = logs
	logger.Logger = logger.Logger.Desugar().WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(c, core)
	})).Sugar()
}

// logsContaining returns the recorded lines whose message contains substr.
// Tests pass a substring unique to them, such as their livestream id.
func logsContaining(substr string) []observer.LoggedEntry {
	return testLogs.FilterMessageSnippet(substr).All()
}

func requireFFmpeg(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found in PATH")
	}
}

// readFLV demuxes testdata/{name} into its packets, in file order.
func readFLV(t *testing.T, name string) []av.Packet {
	t.Helper()

	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("open fixture %s: %v", name, err)
	}
	defer f.Close()

	demuxer := flv.NewDemuxer(f)
	var pkts []av.Packet
	for {
		pkt, err := demuxer.ReadPacket()
		if errors.Is(err, io.EOF) {
			return pkts
		}
		if err != nil {
			t.Fatalf("read fixture %s: %v", name, err)
		}
		pkts = append(pkts, pkt)
	}
}

type storageCall struct {
	Kind         string // "segment" or "thumbnail"
	Index        int    // position among all calls, starting at 0
	StreamID     string
	QualityIndex int
	SourcePath   string
	SavedPath    string // copy of the uploaded file, valid for the whole test
	URL          string
}

// recordingStorage implements storage.Storage. It copies each uploaded file into
// a temp dir so tests can read it after the caller deleted the original.
type recordingStorage struct {
	dir string

	// before runs ahead of each call; a non-nil error fails that call
	// and nothing is copied. It may block to simulate a slow upload.
	before func(call storageCall) error

	mu    sync.Mutex
	calls []storageCall
}

func newRecordingStorage(t *testing.T) *recordingStorage {
	t.Helper()
	return &recordingStorage{dir: t.TempDir()}
}

func (s *recordingStorage) Calls() []storageCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]storageCall(nil), s.calls...)
}

func (s *recordingStorage) AddSegment(ctx context.Context, filePath string, streamId string, qualityIndex int) (string, error) {
	return s.record(storageCall{
		Kind:         "segment",
		StreamID:     streamId,
		QualityIndex: qualityIndex,
		SourcePath:   filePath,
	}, path.Join(streamId, fmt.Sprint(qualityIndex), filepath.Base(filePath)))
}

func (s *recordingStorage) AddThumbnail(ctx context.Context, filePath string, streamId string, contentType string) (string, error) {
	return s.record(storageCall{
		Kind:       "thumbnail",
		StreamID:   streamId,
		SourcePath: filePath,
	}, path.Join(streamId, filepath.Base(filePath)))
}

func (s *recordingStorage) record(call storageCall, objectPath string) (string, error) {
	s.mu.Lock()
	call.Index = len(s.calls)
	s.calls = append(s.calls, call)
	s.mu.Unlock()

	if s.before != nil {
		if err := s.before(call); err != nil {
			return "", err
		}
	}

	saved := filepath.Join(s.dir, filepath.FromSlash(objectPath))
	if err := copyFile(call.SourcePath, saved); err != nil {
		return "", err
	}

	call.SavedPath = saved
	call.URL = recordedURLPrefix + "/" + objectPath

	s.mu.Lock()
	s.calls[call.Index] = call
	s.mu.Unlock()

	return call.URL, nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

func TestFixturesDemux(t *testing.T) {
	cases := []struct {
		name     string
		hasAudio bool
		bFrames  bool
	}{
		{"src_640x360.flv", true, true},
		{"src_360x640.flv", true, true},
		{"src_640x360_noaudio.flv", false, true},
		{"src_1280x720.flv", true, true},
		{"src_1280x720_nobf.flv", true, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var keyframes, audio int
			var bFrames bool
			for _, pkt := range readFLV(t, tc.name) {
				switch pkt.Type {
				case av.H264:
					if pkt.IsKeyFrame {
						keyframes++
					}
					if pkt.CTime != 0 {
						bFrames = true
					}
				case av.AAC:
					audio++
				}
			}

			if keyframes != 3 {
				t.Errorf("keyframes = %d, want 3 (one every 2s over 6s)", keyframes)
			}
			if bFrames != tc.bFrames {
				t.Errorf("fixture has B-frames = %v, want %v", bFrames, tc.bFrames)
			}
			if (audio > 0) != tc.hasAudio {
				t.Errorf("audio packets = %d, hasAudio want %v", audio, tc.hasAudio)
			}
		})
	}
}

func TestRecordingStorage(t *testing.T) {
	st := newRecordingStorage(t)
	src := filepath.Join(t.TempDir(), "seg0.ts")
	if err := os.WriteFile(src, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	url, err := st.AddSegment(context.Background(), src, "abc", 2)
	if err != nil {
		t.Fatalf("AddSegment: %v", err)
	}
	if want := "http://minio.test/bucket/abc/2/seg0.ts"; url != want {
		t.Errorf("url = %q, want %q", url, want)
	}

	calls := st.Calls()
	if len(calls) != 1 || calls[0].Kind != "segment" {
		t.Fatalf("calls = %+v", calls)
	}
	if got, err := os.ReadFile(calls[0].SavedPath); err != nil || string(got) != "data" {
		t.Errorf("saved copy = %q, %v", got, err)
	}

	st.before = func(call storageCall) error {
		if call.Index == 1 {
			return errors.New("boom")
		}
		return nil
	}
	if _, err := st.AddSegment(context.Background(), src, "abc", 2); err == nil {
		t.Error("expected injected error on call 1")
	}
}
