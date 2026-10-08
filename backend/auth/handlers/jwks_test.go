package handlers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sen1or/letslive/auth/config"
	"sen1or/letslive/auth/services"
	"sen1or/letslive/shared/pkg/jwtauth"
)

func TestJWKSHandler_publishesOnlyThePublicKey(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer := jwtauth.NewSigner(key)
	handler := &AuthHandler{jwtService: *services.NewJWTService(nil, config.JWT{}, signer)}

	rec := httptest.NewRecorder()
	handler.JWKSHandler(rec, httptest.NewRequest(http.MethodGet, "/v1/.well-known/jwks.json", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not json: %v", err)
	}
	keys, ok := body["keys"].([]any)
	if !ok || len(keys) != 1 {
		t.Fatalf("expected exactly one key, got %v", body["keys"])
	}
	jwk := keys[0].(map[string]any)
	if jwk["kid"] != jwtauth.KeyID(&key.PublicKey) || jwk["alg"] != "ES256" || jwk["crv"] != "P-256" {
		t.Fatalf("unexpected jwk: %v", jwk)
	}
	if _, leaked := jwk["d"]; leaked {
		t.Fatal("the private scalar must never be published")
	}
}
