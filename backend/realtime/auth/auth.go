package auth

import (
	"net/http"

	"sen1or/letslive/shared/pkg/jwtauth"
	"sen1or/letslive/shared/pkg/realtime"

	"github.com/golang-jwt/jwt/v5"
)

const accessTokenCookie = "ACCESS_TOKEN"

type accessClaims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

// Verifier checks the signature itself because the /realtime gateway route has
// no JWT requirement (anonymous viewers must be able to connect).
type Verifier struct {
	jwt *jwtauth.Verifier
}

// NewVerifier verifies ES256 tokens against the auth service's JWKS at jwksURL.
func NewVerifier(jwksURL string) *Verifier {
	return &Verifier{jwt: jwtauth.NewVerifier(jwksURL)}
}

// UserID returns the user id of a valid ACCESS_TOKEN cookie; ok is false for
// anonymous, expired or forged tokens.
func (v *Verifier) UserID(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(accessTokenCookie)
	if err != nil || cookie.Value == "" {
		return "", false
	}

	var claims accessClaims
	if err := v.jwt.Verify(r.Context(), cookie.Value, &claims); err != nil {
		return "", false
	}

	if !realtime.UserTopic(claims.UserID).Valid() {
		return "", false
	}
	return claims.UserID, true
}
