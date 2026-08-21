package services

import (
	"context"
	"testing"

	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/internal/testutil"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

func TestGetUsersStatuses_ReturnsStringKeyedMap(t *testing.T) {
	userA := uuid.Must(uuid.NewV4())
	userB := uuid.Must(uuid.NewV4())

	userRepo := &testutil.FakeUserRepository{
		GetStatusesByIdsFunc: func(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]domains.UserStatus, *response.Response[any]) {
			return map[uuid.UUID]domains.UserStatus{
				userA: domains.UserStatusDisabled,
				userB: domains.UserStatusNormal,
			}, nil
		},
	}
	s := NewUserService(userRepo, nil, nil, nil, MinIOService{})

	statuses, err := s.GetUsersStatuses(context.Background(), []uuid.UUID{userA, userB})
	if err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}
	if statuses[userA.String()] != "disabled" {
		t.Errorf("statuses[%s] = %q, want %q", userA, statuses[userA.String()], "disabled")
	}
	if statuses[userB.String()] != "normal" {
		t.Errorf("statuses[%s] = %q, want %q", userB, statuses[userB.String()], "normal")
	}
}

func TestGetUsersStatuses_PropagatesRepoError(t *testing.T) {
	wantErr := response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_QUERY, nil, nil, nil)
	userRepo := &testutil.FakeUserRepository{
		GetStatusesByIdsFunc: func(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]domains.UserStatus, *response.Response[any]) {
			return nil, wantErr
		},
	}
	s := NewUserService(userRepo, nil, nil, nil, MinIOService{})

	_, err := s.GetUsersStatuses(context.Background(), []uuid.UUID{uuid.Must(uuid.NewV4())})
	if err != wantErr {
		t.Errorf("err = %+v, want %+v", err, wantErr)
	}
}
