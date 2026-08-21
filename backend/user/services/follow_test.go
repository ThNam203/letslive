package services

import (
	"context"
	"testing"

	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/internal/testutil"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

func TestFollow_RejectsWhenFollowerIsDisabled(t *testing.T) {
	follower := uuid.Must(uuid.NewV4())
	followed := uuid.Must(uuid.NewV4())

	userRepo := &testutil.FakeUserRepository{
		GetByIdFunc: func(ctx context.Context, id uuid.UUID) (*domains.User, *response.Response[any]) {
			if id == follower {
				return &domains.User{Id: id, Status: domains.UserStatusDisabled}, nil
			}
			return &domains.User{Id: id, Status: domains.UserStatusNormal}, nil
		},
	}
	called := false
	followRepo := &testutil.FakeFollowRepository{
		FollowUserFunc: func(ctx context.Context, a, b uuid.UUID) *response.Response[any] {
			called = true
			return nil
		},
	}
	s := NewFollowService(followRepo, userRepo)

	err := s.Follow(context.Background(), follower.String(), followed.String())
	if err == nil {
		t.Fatal("expected an error when the follower is disabled, got nil")
	}
	if called {
		t.Error("FollowUser should not be called when the follower is disabled")
	}
}

func TestFollow_RejectsWhenTargetIsDisabled(t *testing.T) {
	follower := uuid.Must(uuid.NewV4())
	followed := uuid.Must(uuid.NewV4())

	userRepo := &testutil.FakeUserRepository{
		GetByIdFunc: func(ctx context.Context, id uuid.UUID) (*domains.User, *response.Response[any]) {
			if id == followed {
				return &domains.User{Id: id, Status: domains.UserStatusDisabled}, nil
			}
			return &domains.User{Id: id, Status: domains.UserStatusNormal}, nil
		},
	}
	called := false
	followRepo := &testutil.FakeFollowRepository{
		FollowUserFunc: func(ctx context.Context, a, b uuid.UUID) *response.Response[any] {
			called = true
			return nil
		},
	}
	s := NewFollowService(followRepo, userRepo)

	err := s.Follow(context.Background(), follower.String(), followed.String())
	if err == nil {
		t.Fatal("expected an error when the target is disabled, got nil")
	}
	if called {
		t.Error("FollowUser should not be called when the target is disabled")
	}
}

func TestFollow_AllowsWhenBothNormal(t *testing.T) {
	follower := uuid.Must(uuid.NewV4())
	followed := uuid.Must(uuid.NewV4())

	userRepo := &testutil.FakeUserRepository{
		GetByIdFunc: func(ctx context.Context, id uuid.UUID) (*domains.User, *response.Response[any]) {
			return &domains.User{Id: id, Status: domains.UserStatusNormal}, nil
		},
	}
	called := false
	followRepo := &testutil.FakeFollowRepository{
		FollowUserFunc: func(ctx context.Context, a, b uuid.UUID) *response.Response[any] {
			called = true
			return nil
		},
	}
	s := NewFollowService(followRepo, userRepo)

	if err := s.Follow(context.Background(), follower.String(), followed.String()); err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}
	if !called {
		t.Error("FollowUser should be called when both accounts are normal")
	}
}
