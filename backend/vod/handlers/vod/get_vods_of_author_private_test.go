package vod

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"sen1or/letslive/shared/pkg/logger"
)

func TestMain(m *testing.M) {
	logger.Init(logger.LogLevel(logger.Debug))
	os.Exit(m.Run())
}

// the author route used to answer 400 "invalid input" when the cookie was bad
func TestGetVODsOfAuthorPrivateHandler_rejectsMissingCookieAsUnauthorized(t *testing.T) {
	rec := httptest.NewRecorder()

	(&VODHandler{}).GetVODsOfAuthorPrivateHandler(rec, httptest.NewRequest(http.MethodGet, "/v1/vods/author", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	var body struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not json: %v", err)
	}
	if body.Key != "res_err_unauthorized" {
		t.Fatalf("key = %q, want res_err_unauthorized", body.Key)
	}
}

func TestGetVODsOfAuthorPrivateHandler_rejectsForgedCookieAsUnauthorized(t *testing.T) {
	t.Setenv("JWKS_URL", "http://127.0.0.1:1/jwks.json")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/vods/author", nil)
	req.AddCookie(&http.Cookie{Name: "ACCESS_TOKEN", Value: "eyJhbGciOiJFUzI1NiIsImtpZCI6ImEifQ.eyJ1c2VySWQiOiJ4IiwiZXhwIjo5OTk5OTk5OTk5fQ.AAAA"})

	(&VODHandler{}).GetVODsOfAuthorPrivateHandler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}
