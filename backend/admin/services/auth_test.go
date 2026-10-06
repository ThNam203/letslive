package services

import (
	"context"
	"testing"
	"time"

	"sen1or/letslive/admin/domains"
	"sen1or/letslive/admin/response"
	"sen1or/letslive/admin/types"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type fakeAdminAccountRepo struct {
	byEmail map[string]*domains.AdminAccount
	byID    map[uuid.UUID]*domains.AdminAccount
}

func newFakeRepo(accounts ...*domains.AdminAccount) *fakeAdminAccountRepo {
	r := &fakeAdminAccountRepo{
		byEmail: map[string]*domains.AdminAccount{},
		byID:    map[uuid.UUID]*domains.AdminAccount{},
	}
	for _, a := range accounts {
		r.byEmail[a.Email] = a
		r.byID[a.Id] = a
	}
	return r
}

func (r *fakeAdminAccountRepo) GetByEmail(ctx context.Context, email string) (*domains.AdminAccount, *response.Response[any]) {
	if a, ok := r.byEmail[email]; ok {
		return a, nil
	}
	return nil, response.NewResponseFromTemplate[any](response.RES_ERR_ADMIN_NOT_FOUND, nil, nil, nil)
}

func (r *fakeAdminAccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*domains.AdminAccount, *response.Response[any]) {
	if a, ok := r.byID[id]; ok {
		return a, nil
	}
	return nil, response.NewResponseFromTemplate[any](response.RES_ERR_ADMIN_NOT_FOUND, nil, nil, nil)
}

func mustHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %s", err)
	}
	return string(hash)
}

func TestAuthService_Login(t *testing.T) {
	knownAdmin := &domains.AdminAccount{
		Id:           uuid.Must(uuid.NewV4()),
		Email:        "admin@letslive.app",
		PasswordHash: mustHash(t, "correct-password"),
		CreatedAt:    time.Now(),
	}

	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{name: "correct credentials", email: "admin@letslive.app", password: "correct-password", wantErr: false},
		{name: "wrong password", email: "admin@letslive.app", password: "wrong-password", wantErr: true},
		{name: "unknown email", email: "nobody@letslive.app", password: "correct-password", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewAuthService(newFakeRepo(knownAdmin))
			account, errResp := svc.Login(context.Background(), tt.email, tt.password)

			if tt.wantErr {
				if errResp == nil {
					t.Fatalf("expected error, got account %+v", account)
				}
				if errResp.Code != response.RES_ERR_INVALID_CREDENTIALS_CODE {
					t.Fatalf("expected RES_ERR_INVALID_CREDENTIALS_CODE, got %d", errResp.Code)
				}
				return
			}

			if errResp != nil {
				t.Fatalf("unexpected error: %+v", errResp)
			}
			if account.Id != knownAdmin.Id {
				t.Fatalf("got account id %s, want %s", account.Id, knownAdmin.Id)
			}
		})
	}
}

func TestAuthService_GenerateAccessToken(t *testing.T) {
	t.Setenv("ADMIN_JWT_SECRET", "test-secret")

	svc := NewAuthService(newFakeRepo())
	adminId := uuid.Must(uuid.NewV4()).String()

	token, expiresAt, errResp := svc.GenerateAccessToken(adminId)
	if errResp != nil {
		t.Fatalf("unexpected error: %+v", errResp)
	}
	if time.Until(expiresAt) <= 0 || time.Until(expiresAt) > 25*time.Hour {
		t.Fatalf("expiresAt %v not within expected ~24h window", expiresAt)
	}

	claims := &types.AdminClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte("test-secret"), nil
	})
	if err != nil || !parsed.Valid {
		t.Fatalf("generated token does not parse/verify: %s", err)
	}
	if claims.AdminId != adminId {
		t.Fatalf("got adminId %q, want %q", claims.AdminId, adminId)
	}
}
