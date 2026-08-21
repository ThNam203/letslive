package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sen1or/letslive/auth/config"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/dto"
	usergateway "sen1or/letslive/auth/gateway/user"
	"sen1or/letslive/auth/internal/testutil"
	serviceresponse "sen1or/letslive/auth/response"
	"sen1or/letslive/auth/services"
	"sen1or/letslive/auth/types"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const testPassword = "Password123!"

func newTestAuthHandler(t *testing.T, gateway *testutil.FakeUserGateway, authRepo *testutil.FakeAuthRepository, refreshRepo *testutil.FakeRefreshTokenRepository) *AuthHandler {
	t.Helper()
	t.Setenv("ACCESS_TOKEN_SECRET", "test-access-secret")
	t.Setenv("REFRESH_TOKEN_SECRET", "test-refresh-secret")
	t.Setenv("REACTIVATION_TOKEN_SECRET", "test-reactivation-secret")

	jwtCfg := config.JWT{RefreshTokenMaxAge: 3600, AccessTokenMaxAge: 900, Consumer: "test", Issuer: "test", Subject: "test"}

	authService := services.NewAuthService(authRepo, gateway)
	googleAuthService := services.NewGoogleAuthService(authRepo, gateway)
	jwtService := services.NewJWTService(refreshRepo, jwtCfg, gateway)

	return NewAuthHandler(*jwtService, *authService, services.VerificationService{}, *googleAuthService, "")
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %s", err)
	}
	return string(hash)
}

func TestLogInHandler_NormalAccountLogsInSuccessfully(t *testing.T) {
	userId := uuid.Must(uuid.NewV4())
	passwordHash := hashPassword(t, testPassword)

	authRepo := &testutil.FakeAuthRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*domains.Auth, *serviceresponse.Response[any]) {
			return &domains.Auth{UserId: &userId, Email: email, PasswordHash: passwordHash}, nil
		},
	}
	gateway := &testutil.FakeUserGateway{
		GetUserStatusFunc: func(ctx context.Context, id string) (string, *serviceresponse.Response[any]) {
			return usergateway.UserStatusNormal, nil
		},
	}
	refreshRepo := &testutil.FakeRefreshTokenRepository{
		InsertFunc: func(ctx context.Context, token *domains.RefreshToken) *serviceresponse.Response[any] {
			return nil
		},
	}

	h := newTestAuthHandler(t, gateway, authRepo, refreshRepo)

	body, _ := json.Marshal(map[string]string{"email": "user@example.com", "password": testPassword})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("User-Agent", "letslive-mobile-test")
	rec := httptest.NewRecorder()

	h.LogInHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if _, ok := rec.Result().Header["Set-Cookie"]; !ok {
		t.Error("expected Set-Cookie header on successful login")
	}
}

func TestLogInHandler_DisabledAccountReturnsReactivationTokenWithoutCookies(t *testing.T) {
	userId := uuid.Must(uuid.NewV4())
	passwordHash := hashPassword(t, testPassword)

	authRepo := &testutil.FakeAuthRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*domains.Auth, *serviceresponse.Response[any]) {
			return &domains.Auth{UserId: &userId, Email: email, PasswordHash: passwordHash}, nil
		},
	}
	gateway := &testutil.FakeUserGateway{
		GetUserStatusFunc: func(ctx context.Context, id string) (string, *serviceresponse.Response[any]) {
			return usergateway.UserStatusDisabled, nil
		},
	}
	h := newTestAuthHandler(t, gateway, authRepo, &testutil.FakeRefreshTokenRepository{})

	body, _ := json.Marshal(map[string]string{"email": "user@example.com", "password": testPassword})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("User-Agent", "letslive-mobile-test")
	rec := httptest.NewRecorder()

	h.LogInHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	if _, ok := rec.Result().Header["Set-Cookie"]; ok {
		t.Error("expected no Set-Cookie header when account is disabled")
	}

	var res serviceresponse.Response[dto.AccountDisabledResponseDTO]
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response body: %s", err)
	}
	if res.Data == nil || res.Data.ReactivationToken == "" {
		t.Fatalf("expected a non-empty reactivationToken, got %+v", res.Data)
	}
}

func TestReactivateHandler_ValidTokenReactivatesAndLogsIn(t *testing.T) {
	userId := uuid.Must(uuid.NewV4())

	var updatedStatus string
	gateway := &testutil.FakeUserGateway{
		UpdateUserStatusFunc: func(ctx context.Context, id string, status string) *serviceresponse.Response[any] {
			updatedStatus = status
			return nil
		},
	}
	refreshRepo := &testutil.FakeRefreshTokenRepository{
		InsertFunc: func(ctx context.Context, token *domains.RefreshToken) *serviceresponse.Response[any] {
			return nil
		},
	}
	h := newTestAuthHandler(t, gateway, &testutil.FakeAuthRepository{}, refreshRepo)

	token, tokenErr := h.jwtService.GenerateReactivationToken(context.Background(), userId.String())
	if tokenErr != nil {
		t.Fatalf("failed to generate reactivation token: %+v", tokenErr)
	}

	body, _ := json.Marshal(map[string]string{"reactivationToken": token})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/reactivate", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.ReactivateHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if updatedStatus != usergateway.UserStatusNormal {
		t.Errorf("updated status = %q, want %q", updatedStatus, usergateway.UserStatusNormal)
	}
	if _, ok := rec.Result().Header["Set-Cookie"]; !ok {
		t.Error("expected Set-Cookie header after successful reactivation")
	}
}

func TestReactivateHandler_InvalidTokenIsRejectedWithoutUpdatingStatus(t *testing.T) {
	called := false
	gateway := &testutil.FakeUserGateway{
		UpdateUserStatusFunc: func(ctx context.Context, id string, status string) *serviceresponse.Response[any] {
			called = true
			return nil
		},
	}
	h := newTestAuthHandler(t, gateway, &testutil.FakeAuthRepository{}, &testutil.FakeRefreshTokenRepository{})

	body, _ := json.Marshal(map[string]string{"reactivationToken": "not-a-real-token"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/reactivate", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.ReactivateHandler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if called {
		t.Error("UpdateUserStatus should not be called for an invalid token")
	}
}

func TestLogOutAllHandler_ValidAccessTokenRevokesAllSessions(t *testing.T) {
	userId := uuid.Must(uuid.NewV4())
	var revokedUserId uuid.UUID
	refreshRepo := &testutil.FakeRefreshTokenRepository{
		InsertFunc: func(ctx context.Context, token *domains.RefreshToken) *serviceresponse.Response[any] {
			return nil
		},
		RevokeAllTokensOfUserFunc: func(ctx context.Context, id uuid.UUID) *serviceresponse.Response[any] {
			revokedUserId = id
			return nil
		},
	}
	h := newTestAuthHandler(t, &testutil.FakeUserGateway{}, &testutil.FakeAuthRepository{}, refreshRepo)

	pair, genErr := h.jwtService.GenerateTokenPair(context.Background(), userId.String())
	if genErr != nil {
		t.Fatalf("failed to generate token pair: %+v", genErr)
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/auth/logout-all", nil)
	req.AddCookie(&http.Cookie{Name: "ACCESS_TOKEN", Value: pair.AccessToken})
	rec := httptest.NewRecorder()

	h.LogOutAllHandler(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if revokedUserId != userId {
		t.Errorf("revoked userId = %s, want %s", revokedUserId, userId)
	}
	if _, ok := rec.Result().Header["Set-Cookie"]; !ok {
		t.Error("expected Set-Cookie header clearing local cookies")
	}
}

func TestLogOutAllHandler_ForgedCookieIsRejectedWithoutRevoking(t *testing.T) {
	called := false
	refreshRepo := &testutil.FakeRefreshTokenRepository{
		RevokeAllTokensOfUserFunc: func(ctx context.Context, id uuid.UUID) *serviceresponse.Response[any] {
			called = true
			return nil
		},
	}
	h := newTestAuthHandler(t, &testutil.FakeUserGateway{}, &testutil.FakeAuthRepository{}, refreshRepo)

	victimId := uuid.Must(uuid.NewV4())
	claims := types.MyClaims{
		UserId: victimId.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	forgedToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("wrong-secret"))
	if err != nil {
		t.Fatalf("failed to build forged token: %s", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/auth/logout-all", nil)
	req.AddCookie(&http.Cookie{Name: "ACCESS_TOKEN", Value: forgedToken})
	rec := httptest.NewRecorder()

	h.LogOutAllHandler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if called {
		t.Error("RevokeAllTokensOfUser should not be called for a forged cookie")
	}
}

func TestLogOutHandler_RevokesRefreshTokenAndClearsCookies(t *testing.T) {
	var foundValue string
	var updatedToken *domains.RefreshToken
	refreshRepo := &testutil.FakeRefreshTokenRepository{
		FindByValueFunc: func(ctx context.Context, value string) (*domains.RefreshToken, *serviceresponse.Response[any]) {
			foundValue = value
			return &domains.RefreshToken{Token: value}, nil
		},
		UpdateFunc: func(ctx context.Context, token *domains.RefreshToken) *serviceresponse.Response[any] {
			updatedToken = token
			return nil
		},
	}
	h := newTestAuthHandler(t, &testutil.FakeUserGateway{}, &testutil.FakeAuthRepository{}, refreshRepo)

	req := httptest.NewRequest(http.MethodDelete, "/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "REFRESH_TOKEN", Value: "some-refresh-token-value"})
	rec := httptest.NewRecorder()

	h.LogOutHandler(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if foundValue != "some-refresh-token-value" {
		t.Errorf("FindByValue called with %q, want %q", foundValue, "some-refresh-token-value")
	}
	if updatedToken == nil || updatedToken.RevokedAt == nil {
		t.Error("expected the refresh token to be marked revoked")
	}
	if _, ok := rec.Result().Header["Set-Cookie"]; !ok {
		t.Error("expected Set-Cookie header clearing cookies")
	}
}

func TestLogOutHandler_MissingRefreshTokenCookieStillClearsCookies(t *testing.T) {
	called := false
	refreshRepo := &testutil.FakeRefreshTokenRepository{
		FindByValueFunc: func(ctx context.Context, value string) (*domains.RefreshToken, *serviceresponse.Response[any]) {
			called = true
			return nil, nil
		},
	}
	h := newTestAuthHandler(t, &testutil.FakeUserGateway{}, &testutil.FakeAuthRepository{}, refreshRepo)

	req := httptest.NewRequest(http.MethodDelete, "/v1/auth/logout", nil)
	rec := httptest.NewRecorder()

	h.LogOutHandler(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if called {
		t.Error("should not attempt revocation when no refresh token cookie is present")
	}
	if _, ok := rec.Result().Header["Set-Cookie"]; !ok {
		t.Error("expected Set-Cookie header clearing cookies")
	}
}
