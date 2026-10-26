package webserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"sen1or/letslive/shared/pkg/logger"
)

const testStreamID = "123e4567-e89b-42d3-a456-426614174000"

type pathRecorder struct {
	got    string
	status int // 0 answers 204
}

func (p *pathRecorder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.got = r.URL.Path
	if p.status != 0 {
		w.WriteHeader(p.status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// requestLogs records every request log line, at every level.
var requestLogs *observer.ObservedLogs

// TestMain initializes the shared logger, which the logging middleware needs,
// and records its output.
func TestMain(m *testing.M) {
	logger.Init(logger.LogLevel(logger.Error))
	core, logs := observer.New(zapcore.DebugLevel)
	requestLogs = logs
	logger.Logger = logger.Logger.Desugar().WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(c, core)
	})).Sugar()
	os.Exit(m.Run())
}

// loggedFor returns the request log lines for uri.
func loggedFor(uri string) []observer.LoggedEntry {
	var out []observer.LoggedEntry
	for _, e := range requestLogs.FilterFieldKey("uri").All() {
		if e.ContextMap()["uri"] == uri {
			out = append(out, e)
		}
	}
	return out
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

// Live LL-HLS requests come several times a second per viewer, so successful
// ones are not logged; failures and every other route still are.
func TestRequestLogging(t *testing.T) {
	h, rec, base := newTestServer(t)
	if err := os.MkdirAll(filepath.Join(base, "vods", "y"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "vods", "y", "index.m3u8"), []byte("#EXTM3U"), 0o644); err != nil {
		t.Fatal(err)
	}

	serve := func(uri string) {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, uri, nil))
	}

	part := "/static/" + testStreamID + "/source/part1.mp4"
	serve(part)
	if got := loggedFor(part); len(got) != 0 {
		t.Errorf("successful live request logged: %+v", got)
	}

	rec.status = http.StatusOK
	playlist := "/static/" + testStreamID + "/source/stream.m3u8?_HLS_msn=3"
	serve(playlist)
	if got := loggedFor(playlist); len(got) != 0 {
		t.Errorf("successful live playlist request logged: %+v", got)
	}

	rec.status = http.StatusNotFound
	missing := "/static/" + testStreamID + "/index.m3u8"
	serve(missing)
	if got := loggedFor(missing); len(got) != 1 || got[0].Level != zapcore.ErrorLevel {
		t.Errorf("failed live request: logged %+v, want one error line", got)
	}

	vod := "/static/vods/y/index.m3u8"
	serve(vod)
	if got := loggedFor(vod); len(got) != 1 || got[0].Level != zapcore.InfoLevel {
		t.Errorf("VOD request: logged %+v, want the shared middleware's info line", got)
	}
}
