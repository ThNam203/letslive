package userservice

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sen1or/letslive/shared/pkg/discovery"
)

// staticRegistry resolves every service to one fixed address.
type staticRegistry struct {
	discovery.Registry
	addr string
	err  error
}

func (r staticRegistry) ServiceAddress(context.Context, string) (string, error) {
	return r.addr, r.err
}

func TestSendDmPostsTheInternalPushContract(t *testing.T) {
	var gotPath, gotRequestID string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotRequestID = r.Header.Get("X-Request-ID")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
	}))
	defer srv.Close()
	g := NewGateway(staticRegistry{addr: strings.TrimPrefix(srv.URL, "http://")})

	err := g.SendDmPush(context.Background(), []string{"u1", "u2"}, "alice", "hi", "conv1")
	if err != nil {
		t.Fatalf("SendDmPush: %v", err)
	}

	if gotPath != "/v1/internal/push" || gotRequestID == "" {
		t.Fatalf("path = %q, request id = %q", gotPath, gotRequestID)
	}
	if gotBody["title"] != "alice" || gotBody["body"] != "hi" {
		t.Fatalf("body = %v", gotBody)
	}
	data, _ := gotBody["data"].(map[string]any)
	if data["type"] != "dm" || data["referenceId"] != "conv1" || data["actionUrl"] != "/messages/conv1" || len(data) != 3 {
		t.Fatalf("data = %v", data)
	}
	if ids, _ := gotBody["userIds"].([]any); len(ids) != 2 {
		t.Fatalf("userIds = %v", gotBody["userIds"])
	}
}

func TestSendDmReportsFailures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	if err := NewGateway(staticRegistry{addr: strings.TrimPrefix(srv.URL, "http://")}).
		SendDmPush(context.Background(), []string{"u1"}, "a", "b", "c"); err == nil {
		t.Fatal("expected an error for a 400 response")
	}
	if err := NewGateway(staticRegistry{err: errors.New("no instance")}).
		SendDmPush(context.Background(), []string{"u1"}, "a", "b", "c"); err == nil {
		t.Fatal("expected an error when discovery fails")
	}
}
