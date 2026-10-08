package jwtauth

import (
	"context"
	"errors"
	"os"
	"sync"

	"github.com/golang-jwt/jwt/v5"
)

// JWKSURLEnv names the environment variable holding the auth service's JWKS URL.
const JWKSURLEnv = "JWKS_URL"

var (
	defaultMu       sync.Mutex
	defaultVerifier *Verifier
)

// SetDefault installs the verifier Verify uses, e.g. the auth service's own.
func SetDefault(v *Verifier) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultVerifier = v
}

func getDefault() (*Verifier, error) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultVerifier != nil {
		return defaultVerifier, nil
	}
	url := os.Getenv(JWKSURLEnv)
	if url == "" {
		return nil, errors.New(JWKSURLEnv + " is not set")
	}
	defaultVerifier = NewVerifier(url)
	return defaultVerifier, nil
}

// Verify checks tokenString with the default verifier, which reads JWKS_URL
// on first use. It fails closed when the variable is missing.
func Verify(ctx context.Context, tokenString string, claims jwt.Claims) error {
	v, err := getDefault()
	if err != nil {
		return err
	}
	return v.Verify(ctx, tokenString, claims)
}
