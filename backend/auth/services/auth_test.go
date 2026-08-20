package services

import (
	"context"
	"testing"

	usergateway "sen1or/letslive/auth/gateway/user"
	"sen1or/letslive/auth/internal/testutil"
	serviceresponse "sen1or/letslive/auth/response"

	"github.com/gofrs/uuid/v5"
)

func TestGetUserStatus_ReturnsStatusFromGateway(t *testing.T) {
	userId := uuid.Must(uuid.NewV4())
	gateway := &testutil.FakeUserGateway{
		GetUserStatusFunc: func(ctx context.Context, id string) (string, *serviceresponse.Response[any]) {
			if id != userId.String() {
				t.Errorf("gateway called with id = %q, want %q", id, userId.String())
			}
			return "disabled", nil
		},
	}
	s := NewAuthService(&testutil.FakeAuthRepository{}, gateway)

	status, err := s.GetUserStatus(context.Background(), userId)
	if err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}
	if status != "disabled" {
		t.Errorf("status = %q, want %q", status, "disabled")
	}
}

func TestGetUserStatus_PropagatesGatewayError(t *testing.T) {
	wantErr := serviceresponse.NewResponseFromTemplate[any](serviceresponse.RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	gateway := &testutil.FakeUserGateway{
		GetUserStatusFunc: func(ctx context.Context, id string) (string, *serviceresponse.Response[any]) {
			return "", wantErr
		},
	}
	s := NewAuthService(&testutil.FakeAuthRepository{}, gateway)

	_, err := s.GetUserStatus(context.Background(), uuid.Must(uuid.NewV4()))
	if err != wantErr {
		t.Errorf("err = %+v, want %+v", err, wantErr)
	}
}

func TestReactivateUser_SetsStatusToNormal(t *testing.T) {
	userId := uuid.Must(uuid.NewV4())
	var gotStatus string
	gateway := &testutil.FakeUserGateway{
		UpdateUserStatusFunc: func(ctx context.Context, id string, status string) *serviceresponse.Response[any] {
			gotStatus = status
			return nil
		},
	}
	s := NewAuthService(&testutil.FakeAuthRepository{}, gateway)

	if err := s.ReactivateUser(context.Background(), userId.String()); err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}
	if gotStatus != usergateway.UserStatusNormal {
		t.Errorf("status sent = %q, want %q", gotStatus, usergateway.UserStatusNormal)
	}
}
