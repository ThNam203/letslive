package jwtauth

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultRefreshInterval = 10 * time.Minute
	// an unknown kid triggers an early refetch, but not more often than this,
	// so tokens with made-up kids can't make us hammer the auth service
	minRefetchInterval = 30 * time.Second
	fetchTimeout       = 5 * time.Second
	maxJWKSBytes       = 64 << 10
	clockLeeway        = 5 * time.Second
)

var ErrNoKey = errors.New("no signing key available")

// Verifier checks ES256 tokens against a JWKS it fetches and caches. When a
// refresh fails it keeps using the keys it already has.
type Verifier struct {
	url             string
	client          *http.Client
	refreshInterval time.Duration

	// fetchMu lets one JWKS fetch run at a time. mu only guards the fields
	// below and is never held across the fetch, so cached keys never wait on it.
	fetchMu     sync.Mutex
	mu          sync.Mutex
	keys        map[string]*ecdsa.PublicKey
	fetchedAt   time.Time
	lastAttempt time.Time
}

// NewVerifier fetches keys lazily from url on first use.
func NewVerifier(url string) *Verifier {
	return &Verifier{
		url:             url,
		client:          &http.Client{Timeout: fetchTimeout},
		refreshInterval: defaultRefreshInterval,
	}
}

// NewStaticVerifier verifies against fixed keys, for the service that owns
// them and for tests.
func NewStaticVerifier(keys ...*ecdsa.PublicKey) *Verifier {
	v := &Verifier{keys: make(map[string]*ecdsa.PublicKey, len(keys)), fetchedAt: time.Now()}
	for _, key := range keys {
		v.keys[KeyID(key)] = key
	}
	return v
}

// Verify parses and validates tokenString into claims. Only ES256 is accepted,
// an expiry is required, and the token must name a key from the JWKS.
func (v *Verifier) Verify(ctx context.Context, tokenString string, claims jwt.Claims) error {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{Algorithm}),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(clockLeeway),
	)
	_, err := parser.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("token has no kid")
		}
		return v.keyFor(ctx, kid)
	})
	return err
}

func (v *Verifier) keyFor(ctx context.Context, kid string) (*ecdsa.PublicKey, error) {
	key, known, stale := v.cached(kid)
	if known {
		// serve the cached key now and refresh off the request path; a failed
		// refresh keeps the cache, so auth being briefly down is harmless
		if stale && v.fetchMu.TryLock() {
			if v.claimFetch() {
				go func() {
					defer v.fetchMu.Unlock()
					_ = v.refresh(context.Background())
				}()
			} else {
				v.fetchMu.Unlock()
			}
		}
		return key, nil
	}

	// an unknown kid has to wait for the fetch, and for one already in flight
	v.fetchMu.Lock()
	defer v.fetchMu.Unlock()
	if key, known, _ := v.cached(kid); known {
		return key, nil
	}
	if !v.claimFetch() {
		return nil, fmt.Errorf("%w: unknown kid %q", ErrNoKey, kid)
	}
	if err := v.refresh(ctx); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNoKey, err)
	}
	if key, known, _ := v.cached(kid); known {
		return key, nil
	}
	return nil, fmt.Errorf("%w: unknown kid %q", ErrNoKey, kid)
}

func (v *Verifier) cached(kid string) (key *ecdsa.PublicKey, known, stale bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	key, known = v.keys[kid]
	return key, known, time.Since(v.fetchedAt) > v.refreshInterval
}

// claimFetch reports whether a fetch may start now and records the attempt.
func (v *Verifier) claimFetch() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.url == "" || time.Since(v.lastAttempt) < minRefetchInterval {
		return false
	}
	v.lastAttempt = time.Now()
	return true
}

// refresh fetches the JWKS and replaces the cached keys; the caller holds
// v.fetchMu, and v.mu is only taken for the swap.
func (v *Verifier) refresh(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.url, nil)
	if err != nil {
		return err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch jwks: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch jwks: status %d", resp.StatusCode)
	}

	var set JWKS
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxJWKSBytes)).Decode(&set); err != nil {
		return fmt.Errorf("decode jwks: %w", err)
	}

	keys := make(map[string]*ecdsa.PublicKey, len(set.Keys))
	for _, jwk := range set.Keys {
		pub, err := jwk.publicKey()
		if err != nil {
			continue
		}
		// the id is recomputed so a JWKS can't pin a key under another's kid
		keys[KeyID(pub)] = pub
	}
	if len(keys) == 0 {
		return errors.New("jwks has no usable keys")
	}

	v.mu.Lock()
	v.keys = keys
	v.fetchedAt = time.Now()
	v.mu.Unlock()
	return nil
}
