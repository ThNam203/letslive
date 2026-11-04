package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/dto"
	"sen1or/letslive/auth/utils"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/bcrypt"
)

func TestMain(m *testing.M) {
	logger.Init(logger.LogLevel(logger.Debug))
	m.Run()
}

type fakeAuthRepo struct {
	auth      domains.Auth
	updates   []string
	updateErr error
}

func (r *fakeAuthRepo) GetByID(context.Context, uuid.UUID) (*domains.Auth, error) {
	a := r.auth
	return &a, nil
}

func (r *fakeAuthRepo) GetByUserID(context.Context, uuid.UUID) (*domains.Auth, error) {
	a := r.auth
	return &a, nil
}

func (r *fakeAuthRepo) GetByEmail(_ context.Context, email string) (*domains.Auth, error) {
	if email != r.auth.Email {
		return nil, domains.ErrAuthNotFound
	}
	a := r.auth
	return &a, nil
}

func (r *fakeAuthRepo) Create(_ context.Context, auth domains.Auth) (*domains.Auth, error) {
	return &auth, nil
}

func (r *fakeAuthRepo) UpdatePasswordHash(_ context.Context, _ string, hash string) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.updates = append(r.updates, hash)
	r.auth.PasswordHash = hash
	return nil
}

const testEmail = "user@test.local"

func repoWithBcrypt(t *testing.T, password string) *fakeAuthRepo {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	return &fakeAuthRepo{auth: domains.Auth{Id: uuid.Must(uuid.NewV4()), Email: testEmail, PasswordHash: string(hash)}}
}

func login(s *AuthService, password string) error {
	_, err := s.GetUserFromCredentials(context.Background(), dto.LogInRequestDTO{Email: testEmail, Password: password})
	return err
}

func TestLoginUpgradesBcryptToArgon2id(t *testing.T) {
	repo := repoWithBcrypt(t, "Password123!")
	s := NewAuthService(repo, nil)

	if err := login(s, "Password123!"); err != nil {
		t.Fatalf("login with a bcrypt hash: %v", err)
	}
	if len(repo.updates) != 1 || !strings.HasPrefix(repo.updates[0], "$argon2id$") {
		t.Fatalf("hash was not upgraded: %v", repo.updates)
	}

	if err := login(s, "Password123!"); err != nil {
		t.Fatalf("login after the upgrade: %v", err)
	}
	if len(repo.updates) != 1 {
		t.Fatalf("an up-to-date hash was rewritten: %d updates", len(repo.updates))
	}
	if err := login(s, "Password123?"); !errors.Is(err, domains.ErrEmailOrPasswordIncorrect) {
		t.Fatalf("wrong password after the upgrade: err = %v", err)
	}
}

func TestWrongPasswordDoesNotTouchTheHash(t *testing.T) {
	repo := repoWithBcrypt(t, "Password123!")
	before := repo.auth.PasswordHash

	if err := login(NewAuthService(repo, nil), "Password123?"); !errors.Is(err, domains.ErrEmailOrPasswordIncorrect) {
		t.Fatalf("err = %v, want ErrEmailOrPasswordIncorrect", err)
	}
	if len(repo.updates) != 0 || repo.auth.PasswordHash != before {
		t.Fatal("a failed login changed the stored hash")
	}
}

func TestFailedUpgradeStillLogsIn(t *testing.T) {
	repo := repoWithBcrypt(t, "Password123!")
	repo.updateErr = errors.New("database down")

	if err := login(NewAuthService(repo, nil), "Password123!"); err != nil {
		t.Fatalf("login failed because the upgrade failed: %v", err)
	}
}

// The lockout case: bcrypt accepts the right 72 bytes plus extra characters.
// The login must still work, and the real password must keep working after.
func TestBcryptTypoPast72BytesDoesNotLockOut(t *testing.T) {
	password := strings.Repeat("ữ", 24) // exactly 72 bytes
	repo := repoWithBcrypt(t, password)
	s := NewAuthService(repo, nil)

	if err := login(s, password+"x"); err != nil {
		t.Fatalf("bcrypt behaviour changed for a login past 72 bytes: %v", err)
	}
	if len(repo.updates) != 0 {
		t.Fatal("a login past 72 bytes was rehashed")
	}
	if err := login(s, password); err != nil {
		t.Fatalf("the real password stopped working: %v", err)
	}
	if len(repo.updates) != 1 {
		t.Fatalf("the real password did not upgrade the hash: %d updates", len(repo.updates))
	}
	if err := login(s, password+"x"); !errors.Is(err, domains.ErrEmailOrPasswordIncorrect) {
		t.Fatalf("after the upgrade the extra characters must be refused: err = %v", err)
	}
}

func TestChangePasswordStoresArgon2id(t *testing.T) {
	repo := repoWithBcrypt(t, "Password123!")
	s := NewAuthService(repo, nil)

	err := s.UpdatePassword(context.Background(), dto.ChangePasswordRequestDTO{OldPassword: "Password123!", NewPassword: "Ữữ!" + strings.Repeat("ữ", 69)}, uuid.Must(uuid.NewV4()))
	if err != nil {
		t.Fatalf("UpdatePassword: %v", err)
	}
	if match, rehash := utils.VerifyPassword(repo.auth.PasswordHash, "Ữữ!"+strings.Repeat("ữ", 69)); !match || rehash {
		t.Fatalf("new hash: match=%v rehash=%v", match, rehash)
	}

	err = s.UpdatePassword(context.Background(), dto.ChangePasswordRequestDTO{OldPassword: "Password123!", NewPassword: "Another123!"}, uuid.Must(uuid.NewV4()))
	if !errors.Is(err, domains.ErrPasswordNotMatch) {
		t.Fatalf("change with the old password after it changed: err = %v", err)
	}
}

func TestLoginWithoutPasswordHashFails(t *testing.T) {
	repo := &fakeAuthRepo{auth: domains.Auth{Id: uuid.Must(uuid.NewV4()), Email: testEmail}}
	if err := login(NewAuthService(repo, nil), "Password123!"); !errors.Is(err, domains.ErrEmailOrPasswordIncorrect) {
		t.Fatalf("google-only account: err = %v", err)
	}
}
