package webserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"sen1or/letslive/shared/pkg/logger"
)

const testStreamID = "123e4567-e89b-42d3-a456-426614174000"

type pathRecorder struct{ got string }

func (p *pathRecorder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.got = r.URL.Path
	w.WriteHeader(http.StatusNoContent)
}

// TestMain initializes the shared logger, which the logging middleware needs.
func TestMain(m *testing.M) {
	logger.Init(logger.LogLevel(logger.Error))
	os.Exit(m.Run())
}

func newTestServer(t *testing.T) (http.Handler, *pathRecorder, string) {
	t.Helper()
	base := t.TempDir()
	rec := &pathRecorder{}
	return NewWebServer(0, base, rec).routes(), rec, base
}

func TestVODServedFromBaseDirectory(t *testing.T) {
	h, rec, base := newTestServer(t)
	dir := filepath.Join(base, "vods", "x")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.m3u8"), []byte("#EXTM3U"), 0o644); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/static/vods/x/index.m3u8", nil))

	if w.Code != http.StatusOK || w.Body.String() != "#EXTM3U" {
		t.Fatalf("got %d %q, want 200 with the file", w.Code, w.Body.String())
	}
	if rec.got != "" {
		t.Errorf("live handler got %q for a VOD request", rec.got)
	}
}

func TestLiveRequestReachesLiveHandlerWithoutStaticPrefix(t *testing.T) {
	h, rec, _ := newTestServer(t)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/static/"+testStreamID+"/index.m3u8", nil))

	if w.Code != http.StatusNoContent {
		t.Fatalf("got %d, want 204 from the live handler", w.Code)
	}
	if want := "/" + testStreamID + "/index.m3u8"; rec.got != want {
		t.Errorf("live handler got path %q, want %q", rec.got, want)
	}
}

func TestHealth(t *testing.T) {
	h, _, _ := newTestServer(t)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/health", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", w.Code)
	}
}
