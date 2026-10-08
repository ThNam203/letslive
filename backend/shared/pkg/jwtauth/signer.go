package jwtauth

import (
	"crypto/ecdsa"

	"github.com/golang-jwt/jwt/v5"
)

// Signer issues ES256 tokens with the auth service's private key.
type Signer struct {
	key *ecdsa.PrivateKey
	kid string
}

func NewSigner(key *ecdsa.PrivateKey) *Signer {
	return &Signer{key: key, kid: KeyID(&key.PublicKey)}
}

// Sign returns the compact JWT for claims, with the key id in the header so
// verifiers can pick the right key from the JWKS.
func (s *Signer) Sign(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = s.kid
	return token.SignedString(s.key)
}

// PublicKey is the key verifiers need; it is what the JWKS endpoint publishes.
func (s *Signer) PublicKey() *ecdsa.PublicKey { return &s.key.PublicKey }

// JWKS is the document served to the other services and the gateway.
func (s *Signer) JWKS() ([]byte, error) { return MarshalJWKS(s.PublicKey()) }
