package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/internal/testutil"
	"sen1or/letslive/user/response"
	"sen1or/letslive/user/services"

	"github.com/gofrs/uuid/v5"
)

func TestGetUsersStatusesInternalHandler_ReturnsStatusMap(t *testing.T) {
	userA := uuid.Must(uuid.NewV4())
	userRepo := &testutil.FakeUserRepository{
		GetStatusesByIdsFunc: func(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]domains.UserStatus, *response.Response[any]) {
			return map[uuid.UUID]domains.UserStatus{userA: domains.UserStatusDisabled}, nil
		},
	}
	userService := services.NewUserService(userRepo, nil, nil, nil, services.MinIOService{})
	h := NewUserHandler(*userService)

	body, _ := json.Marshal(dto.GetUsersStatusesRequestDTO{UserIds: []uuid.UUID{userA}})
	req := httptest.NewRequest(http.MethodPost, "/v1/internal/users/statuses", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.GetUsersStatusesInternalHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var res response.Response[dto.GetUsersStatusesResponseDTO]
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %s", err)
	}
	if res.Data == nil || res.Data.Statuses[userA.String()] != "disabled" {
		t.Fatalf("expected statuses[%s] = disabled, got %+v", userA, res.Data)
	}
}

func TestGetUsersStatusesInternalHandler_InvalidPayloadReturnsError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/internal/users/statuses", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()

	userService := services.NewUserService(&testutil.FakeUserRepository{}, nil, nil, nil, services.MinIOService{})
	h := NewUserHandler(*userService)

	h.GetUsersStatusesInternalHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}
