package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/transcode/domains"
	dto "sen1or/letslive/transcode/gateway/livestream/dto"
)

func TestMain(m *testing.M) {
	logger.Init(logger.Error)
	os.Exit(m.Run())
}

// fixedRegistry resolves every service to one address.
type fixedRegistry struct{ addr string }

func (r fixedRegistry) Register(context.Context, string, string, string, string, []string) error {
	return nil
}
func (r fixedRegistry) Deregister(context.Context, string, string) error { return nil }
func (r fixedRegistry) ServiceAddresses(context.Context, string) ([]string, error) {
	return []string{r.addr}, nil
}
func (r fixedRegistry) ServiceAddress(context.Context, string) (string, error) { return r.addr, nil }

func gatewayFor(t *testing.T, h http.HandlerFunc) *LivestreamGateway {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return NewLivestreamGateway(fixedRegistry{addr: strings.TrimPrefix(srv.URL, "http://")})
}

func TestEndLivestreamReturnsVODId(t *testing.T) {
	vodId := uuid.Must(uuid.NewV4())
	var got dto.EndLivestreamRequestDTO
	g := gatewayFor(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/internal/livestreams/ls-1/end" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"code":200,"data":{"vodId":"` + vodId.String() + `"}}`))
	})

	id, errRes := g.EndLivestream(context.Background(), "ls-1", dto.EndLivestreamRequestDTO{VODStatus: "processing", Duration: 42})
	if errRes != nil {
		t.Fatalf("EndLivestream error: %+v", errRes)
	}
	if id == nil || *id != vodId {
		t.Fatalf("vod id = %v, want %v", id, vodId)
	}
	if got.VODStatus != "processing" || got.Duration != 42 || got.PlaybackURL != nil {
		t.Errorf("sent %+v, want vodStatus processing, duration 42, no playback URL", got)
	}
}

// An older livestream service answers without data; that is not an error.
func TestEndLivestreamWithoutVODId(t *testing.T) {
	g := gatewayFor(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"code":200}`))
	})
	id, errRes := g.EndLivestream(context.Background(), "ls-1", dto.EndLivestreamRequestDTO{})
	if errRes != nil || id != nil {
		t.Fatalf("EndLivestream = %v, %+v; want nil, nil", id, errRes)
	}
}

func TestEndLivestreamFailure(t *testing.T) {
	g := gatewayFor(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"success":false,"code":409,"message":"already ended"}`))
	})
	id, errRes := g.EndLivestream(context.Background(), "ls-1", dto.EndLivestreamRequestDTO{})
	if errRes == nil || id != nil {
		t.Fatalf("EndLivestream = %v, %+v; want nil id and an error", id, errRes)
	}
	if errRes.Message != "already ended" {
		t.Errorf("message = %q", errRes.Message)
	}
}

func TestUpdateVODStatusHonoursContext(t *testing.T) {
	release := make(chan struct{})
	g := gatewayFor(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	})
	// Registered after the server, so it runs first and srv.Close does not
	// wait on the parked handler.
	t.Cleanup(func() { close(release) })

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := g.UpdateVODStatus(ctx, uuid.Must(uuid.NewV4()).String(), domains.VODStatusReady, "u", "", nil)
	if err == nil {
		t.Fatal("UpdateVODStatus succeeded against a server that never answers")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("UpdateVODStatus took %v, want it bounded by ctx", elapsed)
	}
}
