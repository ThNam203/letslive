package jwtauth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type testClaims struct {
	UserId string `json:"userId"`
	jwt.RegisteredClaims
}

func newKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func validClaims() *testClaims {
	return &testClaims{
		UserId:           "user-1",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}
}

func jwksServer(t *testing.T, signer *Signer, hits *atomic.Int32) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		body, _ := signer.JWKS()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestVerify_acceptsTokenSignedByKeyInJWKS(t *testing.T) {
	signer := NewSigner(newKey(t))
	var hits atomic.Int32
	srv := jwksServer(t, signer, &hits)

	token, err := signer.Sign(validClaims())
	if err != nil {
		t.Fatal(err)
	}

	got := &testClaims{}
	if err := NewVerifier(srv.URL).Verify(context.Background(), token, got); err != nil {
		t.Fatalf("expected token to verify: %v", err)
	}
	if got.UserId != "user-1" {
		t.Fatalf("claims not populated: %+v", got)
	}
}

func TestVerify_cachesKeysBetweenCalls(t *testing.T) {
	signer := NewSigner(newKey(t))
	var hits atomic.Int32
	srv := jwksServer(t, signer, &hits)
	verifier := NewVerifier(srv.URL)
	token, _ := signer.Sign(validClaims())

	for range 3 {
		if err := verifier.Verify(context.Background(), token, &testClaims{}); err != nil {
			t.Fatal(err)
		}
	}
	if hits.Load() != 1 {
		t.Fatalf("expected 1 jwks fetch, got %d", hits.Load())
	}
}

func TestVerify_rejectsTokenSignedByAnotherKey(t *testing.T) {
	trusted := NewSigner(newKey(t))
	attacker := NewSigner(newKey(t))
	var hits atomic.Int32
	srv := jwksServer(t, trusted, &hits)

	token, _ := attacker.Sign(validClaims())
	if err := NewVerifier(srv.URL).Verify(context.Background(), token, &testClaims{}); err == nil {
		t.Fatal("token from an untrusted key must not verify")
	}
}

func TestVerify_rejectsForgedKidWithTrustedKidHeader(t *testing.T) {
	trusted := NewSigner(newKey(t))
	attacker := newKey(t)
	var hits atomic.Int32
	srv := jwksServer(t, trusted, &hits)

	forged := jwt.NewWithClaims(jwt.SigningMethodES256, validClaims())
	forged.Header["kid"] = trusted.kid
	token, _ := forged.SignedString(attacker)

	if err := NewVerifier(srv.URL).Verify(context.Background(), token, &testClaims{}); err == nil {
		t.Fatal("signature from the wrong key must not verify even with the right kid")
	}
}

func TestVerify_rejectsHS256AndNone(t *testing.T) {
	signer := NewSigner(newKey(t))
	verifier := NewStaticVerifier(signer.PublicKey())

	hs := jwt.NewWithClaims(jwt.SigningMethodHS256, validClaims())
	hs.Header["kid"] = signer.kid
	hsToken, _ := hs.SignedString([]byte("access_token_secret"))
	if err := verifier.Verify(context.Background(), hsToken, &testClaims{}); err == nil {
		t.Fatal("HS256 token must be rejected")
	}

	none := jwt.NewWithClaims(jwt.SigningMethodNone, validClaims())
	none.Header["kid"] = signer.kid
	noneToken, _ := none.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err := verifier.Verify(context.Background(), noneToken, &testClaims{}); err == nil {
		t.Fatal("unsigned token must be rejected")
	}
}

func TestVerify_rejectsExpiredAndMissingExpiry(t *testing.T) {
	signer := NewSigner(newKey(t))
	verifier := NewStaticVerifier(signer.PublicKey())

	expired := validClaims()
	expired.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))
	token, _ := signer.Sign(expired)
	if err := verifier.Verify(context.Background(), token, &testClaims{}); err == nil {
		t.Fatal("expired token must be rejected")
	}

	noExp := &testClaims{UserId: "user-1"}
	token, _ = signer.Sign(noExp)
	if err := verifier.Verify(context.Background(), token, &testClaims{}); err == nil {
		t.Fatal("token without exp must be rejected")
	}
}

func TestVerify_rejectsTokenWithoutKid(t *testing.T) {
	key := newKey(t)
	verifier := NewStaticVerifier(&key.PublicKey)
	token, _ := jwt.NewWithClaims(jwt.SigningMethodES256, validClaims()).SignedString(key)
	if err := verifier.Verify(context.Background(), token, &testClaims{}); err == nil {
		t.Fatal("token without kid must be rejected")
	}
}

func TestVerify_keepsCachedKeysWhenJWKSFetchFails(t *testing.T) {
	signer := NewSigner(newKey(t))
	var hits atomic.Int32
	srv := jwksServer(t, signer, &hits)
	verifier := NewVerifier(srv.URL)
	token, _ := signer.Sign(validClaims())

	if err := verifier.Verify(context.Background(), token, &testClaims{}); err != nil {
		t.Fatal(err)
	}
	srv.Close()
	verifier.fetchedAt = time.Now().Add(-time.Hour)
	verifier.lastAttempt = time.Now().Add(-time.Hour)

	if err := verifier.Verify(context.Background(), token, &testClaims{}); err != nil {
		t.Fatalf("expected cached key to keep working while auth is down: %v", err)
	}
}

func TestVerify_failsClosedWhenJWKSNeverReachable(t *testing.T) {
	signer := NewSigner(newKey(t))
	token, _ := signer.Sign(validClaims())
	err := NewVerifier("http://127.0.0.1:1/jwks.json").Verify(context.Background(), token, &testClaims{})
	if err == nil {
		t.Fatal("without any key the token must not verify")
	}
}

func TestParsePrivateKey_acceptsPEMAndBase64AndEscapedNewlines(t *testing.T) {
	key := newKey(t)
	der, _ := x509.MarshalPKCS8PrivateKey(key)
	pemText := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))

	escaped := ""
	for _, r := range pemText {
		if r == '\n' {
			escaped += `\n`
		} else {
			escaped += string(r)
		}
	}

	for name, in := range map[string]string{
		"pem":     pemText,
		"base64":  base64.StdEncoding.EncodeToString([]byte(pemText)),
		"escaped": escaped,
	} {
		got, err := ParsePrivateKey(in)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !got.Equal(key) {
			t.Fatalf("%s: parsed a different key", name)
		}
	}
}

func TestParsePrivateKey_rejectsOtherCurvesAndGarbage(t *testing.T) {
	p384, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	der, _ := x509.MarshalPKCS8PrivateKey(p384)
	pemText := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))

	if _, err := ParsePrivateKey(pemText); err == nil {
		t.Fatal("P-384 must be rejected for ES256")
	}
	if _, err := ParsePrivateKey("not a key"); err == nil {
		t.Fatal("garbage must be rejected")
	}
	if _, err := ParsePrivateKey(""); err == nil {
		t.Fatal("empty must be rejected")
	}
}

func TestKeyID_isStableAndDistinct(t *testing.T) {
	a, b := newKey(t), newKey(t)
	if KeyID(&a.PublicKey) != NewSigner(a).kid {
		t.Fatal("kid must be deterministic")
	}
	if KeyID(&a.PublicKey) == KeyID(&b.PublicKey) {
		t.Fatal("different keys must get different kids")
	}
}
