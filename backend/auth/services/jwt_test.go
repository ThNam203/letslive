package services

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"testing"
	"time"

	"sen1or/letslive/auth/config"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/types"
	"sen1or/letslive/shared/pkg/jwtauth"

	"github.com/golang-jwt/jwt/v5"
)

const testRefreshSecret = "test-refresh-secret"

func newTestJWTService(t *testing.T) *JWTService {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return NewJWTService(nil, config.JWT{AccessTokenMaxAge: 60}, jwtauth.NewSigner(key))
}

func signRefresh(t *testing.T, method jwt.SigningMethod, secret string, exp time.Time) string {
	t.Helper()
	claims := types.MyClaims{
		UserId:           "11111111-1111-1111-1111-111111111111",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(exp)},
	}
	s, err := jwt.NewWithClaims(method, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestRefreshToken_badRefreshTokenIsUnauthorized(t *testing.T) {
	t.Setenv("REFRESH_TOKEN_SECRET", testRefreshSecret)
	tests := []struct {
		name  string
		token string
	}{
		{"expired", signRefresh(t, jwt.SigningMethodHS256, testRefreshSecret, time.Now().Add(-time.Hour))},
		{"signed with the wrong key", signRefresh(t, jwt.SigningMethodHS256, "another-secret", time.Now().Add(time.Hour))},
		{"unexpected signing method", signRefresh(t, jwt.SigningMethodHS512, testRefreshSecret, time.Now().Add(time.Hour))},
		{"garbage", "not-a-jwt"},
	}
	svc := newTestJWTService(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.RefreshToken(context.Background(), tt.token)
			if !errors.Is(err, domains.ErrUnauthorized) {
				t.Fatalf("err = %v, want ErrUnauthorized", err)
			}
		})
	}
}

func TestRefreshToken_validTokenYieldsAccessToken(t *testing.T) {
	t.Setenv("REFRESH_TOKEN_SECRET", testRefreshSecret)
	svc := newTestJWTService(t)
	info, err := svc.RefreshToken(context.Background(), signRefresh(t, jwt.SigningMethodHS256, testRefreshSecret, time.Now().Add(time.Hour)))
	if err != nil || info.AccessToken == "" {
		t.Fatalf("got %v, %v", info, err)
	}
}

func TestRefreshToken_missingSecretIsInternal(t *testing.T) {
	t.Setenv("REFRESH_TOKEN_SECRET", "")
	svc := newTestJWTService(t)
	_, err := svc.RefreshToken(context.Background(), signRefresh(t, jwt.SigningMethodHS256, testRefreshSecret, time.Now().Add(time.Hour)))
	if !errors.Is(err, domains.ErrInternal) {
		t.Fatalf("err = %v, want ErrInternal", err)
	}
}
