package utils

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sen1or/letslive/user/response"
)

const testMaxFileBytes = 1 << 20

func newUploadRequest(t *testing.T, field string, size int) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(field, strings.Repeat("a", 200)+".png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(make([]byte, size)); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, "/v1/user/me/profile-picture", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// the http server removes multipart temp files after a handler returns;
	// calling the parser directly skips that
	t.Cleanup(func() {
		if req.MultipartForm != nil {
			req.MultipartForm.RemoveAll()
		}
	})
	return req
}

func TestParseUploadedFileAcceptsFileAtTheLimit(t *testing.T) {
	req := newUploadRequest(t, "profile-picture", testMaxFileBytes)

	file, header, errRes := ParseUploadedFile(httptest.NewRecorder(), req, "profile-picture", testMaxFileBytes)
	if errRes != nil {
		t.Fatalf("got error response %q, want none", errRes.Key)
	}
	defer file.Close()

	if header.Size != testMaxFileBytes {
		t.Errorf("size = %d, want %d", header.Size, testMaxFileBytes)
	}
}

func TestParseUploadedFileRejectsOversizedFiles(t *testing.T) {
	cases := map[string]int{
		"one byte over":             testMaxFileBytes + 1,
		"beyond the request budget": testMaxFileBytes + 2*multipartOverheadBytes,
	}

	for name, size := range cases {
		t.Run(name, func(t *testing.T) {
			req := newUploadRequest(t, "profile-picture", size)

			_, _, errRes := ParseUploadedFile(httptest.NewRecorder(), req, "profile-picture", testMaxFileBytes)
			if errRes == nil {
				t.Fatal("got no error response, want image too large")
			}
			if errRes.Key != response.RES_ERR_IMAGE_TOO_LARGE_KEY {
				t.Errorf("key = %q, want %q", errRes.Key, response.RES_ERR_IMAGE_TOO_LARGE_KEY)
			}
		})
	}
}

func TestParseUploadedFileRejectsMissingField(t *testing.T) {
	req := newUploadRequest(t, "something-else", 10)

	_, _, errRes := ParseUploadedFile(httptest.NewRecorder(), req, "profile-picture", testMaxFileBytes)
	if errRes == nil {
		t.Fatal("got no error response, want invalid payload")
	}
	if errRes.Key != response.RES_ERR_INVALID_PAYLOAD_KEY {
		t.Errorf("key = %q, want %q", errRes.Key, response.RES_ERR_INVALID_PAYLOAD_KEY)
	}
}
