// Package jwtauth signs and verifies the ES256 access tokens.
//
// The auth service holds the private key and publishes the public one as a
// JWKS; every other service only needs the JWKS URL to verify tokens.
package jwtauth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// Algorithm is the only signing method this package accepts.
const Algorithm = "ES256"

// ParsePrivateKey reads an EC P-256 private key from PEM text (PKCS#8 or SEC1).
// The value may be the PEM itself (with real or escaped "\n" line breaks) or
// the PEM base64-encoded, which fits on one line of an .env file.
func ParsePrivateKey(value string) (*ecdsa.PrivateKey, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, errors.New("private key is empty")
	}
	if !strings.HasPrefix(value, "-----BEGIN") {
		decoded, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, fmt.Errorf("private key is neither PEM nor base64: %w", err)
		}
		value = string(decoded)
	}
	value = strings.ReplaceAll(value, `\n`, "\n")

	block, _ := pem.Decode([]byte(value))
	if block == nil {
		return nil, errors.New("private key has no PEM block")
	}

	var key *ecdsa.PrivateKey
	if parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		ecKey, ok := parsed.(*ecdsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is not an EC key")
		}
		key = ecKey
	} else if ecKey, ecErr := x509.ParseECPrivateKey(block.Bytes); ecErr == nil {
		key = ecKey
	} else {
		return nil, fmt.Errorf("private key is neither PKCS#8 nor SEC1: %w", err)
	}

	if key.Curve != elliptic.P256() {
		return nil, errors.New("private key must be on curve P-256 for ES256")
	}
	return key, nil
}

// JWK is a public EC key in JSON Web Key form (RFC 7517).
type JWK struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
}

// JWKS is a JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// coordinate pads to the 32 bytes P-256 coordinates are encoded with.
func coordinate(n *big.Int) string {
	return base64.RawURLEncoding.EncodeToString(n.FillBytes(make([]byte, 32)))
}

// KeyID is the RFC 7638 thumbprint of the public key, so it changes exactly
// when the key does.
func KeyID(pub *ecdsa.PublicKey) string {
	canonical := fmt.Sprintf(`{"crv":"P-256","kty":"EC","x":"%s","y":"%s"}`, coordinate(pub.X), coordinate(pub.Y))
	sum := sha256.Sum256([]byte(canonical))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// NewJWK describes a public key for the JWKS document.
func NewJWK(pub *ecdsa.PublicKey) JWK {
	return JWK{
		Kty: "EC",
		Crv: "P-256",
		X:   coordinate(pub.X),
		Y:   coordinate(pub.Y),
		Kid: KeyID(pub),
		Use: "sig",
		Alg: Algorithm,
	}
}

// MarshalJWKS renders the JWKS document for the given public keys.
func MarshalJWKS(keys ...*ecdsa.PublicKey) ([]byte, error) {
	set := JWKS{Keys: make([]JWK, 0, len(keys))}
	for _, key := range keys {
		set.Keys = append(set.Keys, NewJWK(key))
	}
	return json.Marshal(set)
}

// publicKey rebuilds the key a JWK describes, rejecting anything but P-256.
func (k JWK) publicKey() (*ecdsa.PublicKey, error) {
	if k.Kty != "EC" || k.Crv != "P-256" {
		return nil, fmt.Errorf("unsupported key %s/%s", k.Kty, k.Crv)
	}
	x, err := base64.RawURLEncoding.DecodeString(k.X)
	if err != nil {
		return nil, fmt.Errorf("bad x coordinate: %w", err)
	}
	y, err := base64.RawURLEncoding.DecodeString(k.Y)
	if err != nil {
		return nil, fmt.Errorf("bad y coordinate: %w", err)
	}
	pub := &ecdsa.PublicKey{Curve: elliptic.P256(), X: new(big.Int).SetBytes(x), Y: new(big.Int).SetBytes(y)}
	if !pub.Curve.IsOnCurve(pub.X, pub.Y) {
		return nil, errors.New("point is not on the curve")
	}
	return pub, nil
}
