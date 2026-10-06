package auth

import (
	"net/http"

	"sen1or/letslive/shared/pkg/realtime"

	"github.com/golang-jwt/jwt/v5"
)

const accessTokenCookie = "ACCESS_TOKEN"

type accessClaims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

// Verifier checks the signature itself because the /realtime Kong route has
// no JWT plugin (anonymous viewers must be able to connect).
type Verifier struct {
	secret []byte
}

func NewVerifier(secret string) *Verifier {
	return &Verifier{secret: []byte(secret)}
}

// UserID returns the user id of a valid ACCESS_TOKEN cookie; ok is false for
// anonymous, expired or forged tokens.
func (v *Verifier) UserID(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(accessTokenCookie)
	if err != nil || cookie.Value == "" {
		return "", false
	}

	var claims accessClaims
	_, err = jwt.ParseWithClaims(
		cookie.Value,
		&claims,
		func(*jwt.Token) (any, error) { return v.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return "", false
	}

	if !realtime.UserTopic(claims.UserID).Valid() {
		return "", false
	}
	return claims.UserID, true
}
