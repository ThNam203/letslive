package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"

	"github.com/gofrs/uuid/v5"
)

type fakeDeviceRepo struct {
	byToken map[string]domains.DeviceToken
}

func newFakeDeviceRepo() *fakeDeviceRepo {
	return &fakeDeviceRepo{byToken: map[string]domains.DeviceToken{}}
}

func (f *fakeDeviceRepo) Upsert(_ context.Context, t domains.DeviceToken) (*domains.DeviceToken, error) {
	f.byToken[t.Token] = t
	return &t, nil
}

func (f *fakeDeviceRepo) ListTokensByUserID(_ context.Context, userID uuid.UUID) ([]domains.DeviceToken, error) {
	var out []domains.DeviceToken
	for _, t := range f.byToken {
		if t.UserId == userID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (f *fakeDeviceRepo) DeleteByToken(_ context.Context, token string) error {
	delete(f.byToken, token)
	return nil
}

func (f *fakeDeviceRepo) DeleteByUserAndToken(_ context.Context, userID uuid.UUID, token string) error {
	if t, ok := f.byToken[token]; ok && t.UserId == userID {
		delete(f.byToken, token)
	}
	return nil
}

func newUser() uuid.UUID { return uuid.Must(uuid.NewV4()) }

func TestRegisteringSameTokenForAnotherUserReassignsIt(t *testing.T) {
	repo := newFakeDeviceRepo()
	svc := NewDeviceService(repo)
	ctx := context.Background()
	userA, userB := newUser(), newUser()
	req := dto.RegisterDeviceRequestDTO{Token: "t1", Platform: "ios"}

	if _, err := svc.Register(ctx, userA, req); err != nil {
		t.Fatal(err)
	}
	saved, err := svc.Register(ctx, userB, req)
	if err != nil {
		t.Fatal(err)
	}

	if saved.UserId != userB {
		t.Fatalf("token owner = %s, want %s", saved.UserId, userB)
	}
	if got, _ := repo.ListTokensByUserID(ctx, userA); len(got) != 0 {
		t.Errorf("user A still owns the token: %+v", got)
	}
}

func TestDeleteRemovesOnlyCallersToken(t *testing.T) {
	repo := newFakeDeviceRepo()
	svc := NewDeviceService(repo)
	ctx := context.Background()
	owner, other := newUser(), newUser()
	if _, err := svc.Register(ctx, owner, dto.RegisterDeviceRequestDTO{Token: "t1", Platform: "android"}); err != nil {
		t.Fatal(err)
	}

	if err := svc.Unregister(ctx, other, "t1"); err != nil {
		t.Fatalf("foreign delete must succeed as no-op: %v", err)
	}
	if got, _ := repo.ListTokensByUserID(ctx, owner); len(got) != 1 {
		t.Fatalf("foreign delete removed the token")
	}
	if err := svc.Unregister(ctx, owner, "t1"); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.ListTokensByUserID(ctx, owner); len(got) != 0 {
		t.Fatalf("owner delete did not remove the token")
	}
	if err := svc.Unregister(ctx, owner, "missing"); err != nil {
		t.Fatalf("deleting a missing token must succeed: %v", err)
	}
}

func TestUnregisterRejectsEmptyOrOversizedToken(t *testing.T) {
	svc := NewDeviceService(newFakeDeviceRepo())
	for _, tok := range []string{"", strings.Repeat("a", dto.MaxDeviceTokenLength+1)} {
		if err := svc.Unregister(context.Background(), newUser(), tok); !errors.Is(err, domains.ErrInvalidInput) {
			t.Errorf("token len %d: err = %v, want ErrInvalidInput", len(tok), err)
		}
	}
}
