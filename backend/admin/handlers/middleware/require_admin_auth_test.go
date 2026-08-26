package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sen1or/letslive/admin/types"

	"github.com/golang-jwt/jwt/v5"
)

func signToken(t *testing.T, secret string, adminId string, expiresAt time.Time) string {
	t.Helper()
	claims := types.AdminClaims{
		AdminId: adminId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign test token: %s", err)
	}
	return token
}

func TestRequireAdminAuth(t *testing.T) {
	t.Setenv("ADMIN_JWT_SECRET", "correct-secret")

	tests := []struct {
		name       string
		cookie     *http.Cookie
		wantStatus int
		wantNext   bool
	}{
		{
			name:       "valid token",
			cookie:     &http.Cookie{Name: "ADMIN_ACCESS_TOKEN", Value: signToken(t, "correct-secret", "admin-1", time.Now().Add(time.Hour))},
			wantStatus: http.StatusOK,
			wantNext:   true,
		},
		{
			name:       "expired token",
			cookie:     &http.Cookie{Name: "ADMIN_ACCESS_TOKEN", Value: signToken(t, "correct-secret", "admin-1", time.Now().Add(-time.Hour))},
			wantStatus: http.StatusUnauthorized,
			wantNext:   false,
		},
		{
			name:       "wrong secret (e.g. a main-app token)",
			cookie:     &http.Cookie{Name: "ADMIN_ACCESS_TOKEN", Value: signToken(t, "some-other-secret", "admin-1", time.Now().Add(time.Hour))},
			wantStatus: http.StatusUnauthorized,
			wantNext:   false,
		},
		{
			name:       "malformed token",
			cookie:     &http.Cookie{Name: "ADMIN_ACCESS_TOKEN", Value: "not-a-jwt"},
			wantStatus: http.StatusUnauthorized,
			wantNext:   false,
		},
		{
			name:       "missing cookie",
			cookie:     nil,
			wantStatus: http.StatusUnauthorized,
			wantNext:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			}

			req := httptest.NewRequest(http.MethodGet, "/v1/admin/me", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			rec := httptest.NewRecorder()

			RequireAdminAuth(next)(rec, req)

			if nextCalled != tt.wantNext {
				t.Fatalf("next called = %v, want %v", nextCalled, tt.wantNext)
			}
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
