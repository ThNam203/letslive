package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"sen1or/letslive/admin/response"
	"sen1or/letslive/admin/types"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const AdminIdContextKey contextKey = "adminId"

// RequireAdminAuth does full JWT signature verification against ADMIN_JWT_SECRET —
// unlike the main app's services, which trust Kong's jwt plugin and only decode
// claims unverified. Kong has no jwt plugin on /admin/* routes (see plan/spec), so
// this middleware IS the auth boundary.
func RequireAdminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("ADMIN_ACCESS_TOKEN")
		if err != nil || len(cookie.Value) == 0 {
			writeUnauthorized(w, r)
			return
		}

		claims := types.AdminClaims{}
		token, err := jwt.ParseWithClaims(cookie.Value, &claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(os.Getenv("ADMIN_JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			writeUnauthorized(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), AdminIdContextKey, claims.AdminId)
		next(w, r.WithContext(ctx))
	}
}

func writeUnauthorized(w http.ResponseWriter, r *http.Request) {
	res := response.NewResponseFromTemplate[any](response.RES_ERR_UNAUTHORIZED, nil, nil, nil)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(res.StatusCode)
	_ = json.NewEncoder(w).Encode(res)
}
