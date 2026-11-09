package vod

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sen1or/letslive/vod/domains"
)

func TestCreateVODInternalStatus(t *testing.T) {
	tests := []struct {
		raw    string
		want   domains.VODStatus
		wantOK bool
	}{
		{raw: "", want: domains.VODStatusReady, wantOK: true},
		{raw: "ready", want: domains.VODStatusReady, wantOK: true},
		{raw: "processing", want: domains.VODStatusProcessing, wantOK: true},
		{raw: "failed", wantOK: false},
		{raw: "uploading", wantOK: false},
		{raw: "READY", wantOK: false},
	}
	for _, tt := range tests {
		got, ok := createVODInternalStatus(tt.raw)
		if ok != tt.wantOK || (ok && got != tt.want) {
			t.Errorf("createVODInternalStatus(%q) = %q, %v; want %q, %v", tt.raw, got, ok, tt.want, tt.wantOK)
		}
	}
}

// An unknown status is rejected before the service is reached.
func TestCreateVODInternalHandler_rejectsUnknownStatus(t *testing.T) {
	body := `{"livestreamId":"9b2f8f8e-5b0e-4b8e-9f43-0c1d3f1e2a10","userId":"0d7c1f0e-3a5b-4c6d-8e9f-1a2b3c4d5e6f","title":"t","status":"failed"}`
	rec := httptest.NewRecorder()

	(&VODHandler{}).CreateVODInternalHandler(rec, httptest.NewRequest(http.MethodPost, "/v1/internal/vods", strings.NewReader(body)))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
