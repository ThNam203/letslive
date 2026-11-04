package handlers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sen1or/letslive/auth/config"
	"sen1or/letslive/auth/services"
	"sen1or/letslive/auth/types"
	"sen1or/letslive/shared/pkg/jwtauth"

	"github.com/golang-jwt/jwt/v5"
)

func TestRefreshTokenHandler_badTokensAre401(t *testing.T) {
	const secret = "handler-refresh-secret"
	t.Setenv("REFRESH_TOKEN_SECRET", secret)
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	h := &AuthHandler{jwtService: *services.NewJWTService(nil, config.JWT{}, jwtauth.NewSigner(key))}

	sign := func(signingSecret string, exp time.Time) string {
		s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, types.MyClaims{
			UserId:           "u",
			RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(exp)},
		}).SignedString([]byte(signingSecret))
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	tests := map[string]string{
		"expired":     sign(secret, time.Now().Add(-time.Hour)),
		"wrong key":   sign("other", time.Now().Add(time.Hour)),
		"garbage":     "garbage",
		"empty value": "",
	}
	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh-token", nil)
			req.AddCookie(&http.Cookie{Name: "REFRESH_TOKEN", Value: value})
			rec := httptest.NewRecorder()
			h.RefreshTokenHandler(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401 (%s)", rec.Code, rec.Body.String())
			}
		})
	}
}
