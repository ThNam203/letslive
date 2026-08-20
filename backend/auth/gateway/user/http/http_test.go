package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sen1or/letslive/auth/gateway/user/dto"
	serviceresponse "sen1or/letslive/auth/response"
	"sen1or/letslive/shared/pkg/logger"
)

func TestMain(m *testing.M) {
	logger.Init(logger.Debug)
	m.Run()
}

type fakeRegistry struct {
	addr string
}

func (f *fakeRegistry) Register(ctx context.Context, hostPort string, serviceHealthCheckURL string, serviceName string, instanceId string, tags []string) error {
	return nil
}

func (f *fakeRegistry) Deregister(ctx context.Context, serviceName string, instanceId string) error {
	return nil
}

func (f *fakeRegistry) ServiceAddresses(ctx context.Context, serviceName string) ([]string, error) {
	return []string{f.addr}, nil
}

func (f *fakeRegistry) ServiceAddress(ctx context.Context, serviceName string) (string, error) {
	return f.addr, nil
}

func TestGetUserStatus_ReturnsStatusOnSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/user/user-123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(serviceresponse.Response[dto.GetUserStatusResponseDTO]{
			Success: true,
			Data:    &dto.GetUserStatusResponseDTO{Status: "disabled"},
		})
	}))
	defer server.Close()

	g := NewUserGateway(&fakeRegistry{addr: strings.TrimPrefix(server.URL, "http://")})

	status, err := g.GetUserStatus(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}
	if status != "disabled" {
		t.Errorf("status = %q, want %q", status, "disabled")
	}
}

func TestGetUserStatus_PropagatesNonSuccessResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(serviceresponse.Response[any]{
			Success: false,
			Code:    30000,
			Key:     "res_err_user_not_found",
		})
	}))
	defer server.Close()

	g := NewUserGateway(&fakeRegistry{addr: strings.TrimPrefix(server.URL, "http://")})

	_, err := g.GetUserStatus(context.Background(), "missing-user")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Code != 30000 {
		t.Errorf("err.Code = %d, want 30000", err.Code)
	}
}

func TestUpdateUserStatus_SendsStatusAndSucceeds(t *testing.T) {
	var receivedBody dto.UpdateUserStatusRequestDTO
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/v1/user/user-123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(serviceresponse.Response[any]{Success: true})
	}))
	defer server.Close()

	g := NewUserGateway(&fakeRegistry{addr: strings.TrimPrefix(server.URL, "http://")})

	err := g.UpdateUserStatus(context.Background(), "user-123", "normal")
	if err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}
	if receivedBody.Status != "normal" {
		t.Errorf("sent status = %q, want %q", receivedBody.Status, "normal")
	}
}
