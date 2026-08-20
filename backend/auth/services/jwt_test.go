package services

import (
	"context"
	"testing"
	"time"

	"sen1or/letslive/auth/config"
	"sen1or/letslive/auth/domains"
	usergateway "sen1or/letslive/auth/gateway/user"
	"sen1or/letslive/auth/internal/testutil"
	serviceresponse "sen1or/letslive/auth/response"
	"sen1or/letslive/auth/types"

	"github.com/golang-jwt/jwt/v5"
)

func newTestJWTService(t *testing.T, gateway usergateway.UserGateway) *JWTService {
	t.Helper()
	t.Setenv("REACTIVATION_TOKEN_SECRET", "test-reactivation-secret")
	t.Setenv("ACCESS_TOKEN_SECRET", "test-access-secret")
	t.Setenv("REFRESH_TOKEN_SECRET", "test-refresh-secret")

	repo := &testutil.FakeRefreshTokenRepository{
		InsertFunc: func(ctx context.Context, token *domains.RefreshToken) *serviceresponse.Response[any] {
			return nil
		},
	}
	cfg := config.JWT{RefreshTokenMaxAge: 3600, AccessTokenMaxAge: 900, Consumer: "test", Issuer: "test", Subject: "test"}
	return NewJWTService(repo, cfg, gateway)
}

func TestGenerateAndVerifyReactivationToken_RoundTrip(t *testing.T) {
	s := newTestJWTService(t, &testutil.FakeUserGateway{})

	token, genErr := s.GenerateReactivationToken(context.Background(), "user-123")
	if genErr != nil {
		t.Fatalf("expected no error, got %+v", genErr)
	}

	userId, verifyErr := s.VerifyReactivationToken(context.Background(), token)
	if verifyErr != nil {
		t.Fatalf("expected no error, got %+v", verifyErr)
	}
	if userId != "user-123" {
		t.Errorf("userId = %q, want %q", userId, "user-123")
	}
}

func TestVerifyReactivationToken_RejectsExpiredToken(t *testing.T) {
	s := newTestJWTService(t, &testutil.FakeUserGateway{})

	expiredClaims := types.MyClaims{
		UserId: "user-123",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	expiredToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).SignedString([]byte("test-reactivation-secret"))
	if err != nil {
		t.Fatalf("failed to build expired token: %s", err)
	}

	if _, verifyErr := s.VerifyReactivationToken(context.Background(), expiredToken); verifyErr == nil {
		t.Fatal("expected an error for an expired token, got nil")
	}
}

func TestVerifyReactivationToken_RejectsTokenSignedWithWrongSecret(t *testing.T) {
	s := newTestJWTService(t, &testutil.FakeUserGateway{})

	claims := types.MyClaims{
		UserId: "user-123",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
		},
	}
	wrongSecretToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("some-other-secret"))
	if err != nil {
		t.Fatalf("failed to build token: %s", err)
	}

	if _, verifyErr := s.VerifyReactivationToken(context.Background(), wrongSecretToken); verifyErr == nil {
		t.Fatal("expected an error for a token signed with the wrong secret, got nil")
	}
}

func TestVerifyReactivationToken_RejectsTokenMissingReactivationPurpose(t *testing.T) {
	s := newTestJWTService(t, &testutil.FakeUserGateway{})

	claims := types.MyClaims{
		UserId: "user-123",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
		},
	}
	accessLikeToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-reactivation-secret"))
	if err != nil {
		t.Fatalf("failed to build token: %s", err)
	}

	if _, verifyErr := s.VerifyReactivationToken(context.Background(), accessLikeToken); verifyErr == nil {
		t.Fatal("expected an error for a token missing the reactivation purpose, got nil")
	}
}

func TestVerifyReactivationToken_RejectsRealAccessToken(t *testing.T) {
	s := newTestJWTService(t, &testutil.FakeUserGateway{})

	pair, genErr := s.GenerateTokenPair(context.Background(), "user-123")
	if genErr != nil {
		t.Fatalf("failed to generate token pair: %+v", genErr)
	}

	if _, verifyErr := s.VerifyReactivationToken(context.Background(), pair.AccessToken); verifyErr == nil {
		t.Fatal("expected an error for a real access token presented at the reactivate endpoint, got nil")
	}
}

func TestRefreshToken_RejectsDisabledAccount(t *testing.T) {
	gateway := &testutil.FakeUserGateway{
		GetUserStatusFunc: func(ctx context.Context, id string) (string, *serviceresponse.Response[any]) {
			return usergateway.UserStatusDisabled, nil
		},
	}
	s := newTestJWTService(t, gateway)

	pair, genErr := s.GenerateTokenPair(context.Background(), "user-123")
	if genErr != nil {
		t.Fatalf("failed to generate token pair: %+v", genErr)
	}

	if _, refreshErr := s.RefreshToken(context.Background(), pair.RefreshToken); refreshErr == nil {
		t.Fatal("expected refresh to be rejected for a disabled account, got nil error")
	}
}

func TestRefreshToken_AllowsNormalAccount(t *testing.T) {
	gateway := &testutil.FakeUserGateway{
		GetUserStatusFunc: func(ctx context.Context, id string) (string, *serviceresponse.Response[any]) {
			return usergateway.UserStatusNormal, nil
		},
	}
	s := newTestJWTService(t, gateway)

	pair, genErr := s.GenerateTokenPair(context.Background(), "user-123")
	if genErr != nil {
		t.Fatalf("failed to generate token pair: %+v", genErr)
	}

	accessInfo, refreshErr := s.RefreshToken(context.Background(), pair.RefreshToken)
	if refreshErr != nil {
		t.Fatalf("expected no error, got %+v", refreshErr)
	}
	if accessInfo.AccessToken == "" {
		t.Error("expected a non-empty access token")
	}
}
