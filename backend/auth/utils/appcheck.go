package utils

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/appcheck"

	"sen1or/letslive/shared/pkg/logger"
)

// FirebaseProjectIDEnv turns Firebase App Check verification on. App Check
// tokens are verified against Google's public JWKS, so the project id is the
// only configuration needed (no service account).
const FirebaseProjectIDEnv = "FIREBASE_PROJECT_ID"

const (
	appCheckInitBackoffMin = time.Second
	appCheckInitBackoffMax = time.Minute
)

// Firebase project ids: 6-30 chars, lowercase letters, digits and hyphens,
// starting with a letter and not ending with a hyphen.
var firebaseProjectIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{4,28}[a-z0-9]$`)

// AppCheckVerifier verifies a Firebase App Check token. A nil error means the
// token is genuine, unexpired, issued for this project and not used before.
type AppCheckVerifier interface {
	Verify(ctx context.Context, token string) error
}

// tokenVerifier is the part of *appcheck.Client we use; it lets tests avoid the network.
type tokenVerifier interface {
	VerifyToken(token string) (*appcheck.DecodedAppCheckToken, error)
}

type firebaseAppCheckVerifier struct {
	client tokenVerifier
	guard  ReplayGuard
}

// NewAppCheckVerifierFromEnv returns (nil, nil) when FIREBASE_PROJECT_ID is
// unset or blank: App Check is then off. A malformed project id is an error
// (the caller must stop). Otherwise it returns immediately with a verifier
// that initialises the Firebase SDK in the background, retrying with backoff;
// until it is ready Verify rejects every token, so requests that carry an App
// Check header fail the bot check while header-less (Turnstile) requests are
// unaffected. A JWKS outage at boot therefore never takes the service down.
func NewAppCheckVerifierFromEnv(ctx context.Context) (AppCheckVerifier, error) {
	projectID := strings.TrimSpace(os.Getenv(FirebaseProjectIDEnv))
	if projectID == "" {
		return nil, nil
	}
	if !firebaseProjectIDPattern.MatchString(projectID) {
		return nil, fmt.Errorf("%s %q is not a valid Firebase project id", FirebaseProjectIDEnv, projectID)
	}

	guard := NewMemoryReplayGuard(DefaultReplayGuardMaxEntries)
	build := func(ctx context.Context) (AppCheckVerifier, error) {
		return NewFirebaseAppCheckVerifier(ctx, projectID, guard)
	}
	return newLazyAppCheckVerifier(ctx, build, appCheckInitBackoffMin, appCheckInitBackoffMax), nil
}

// NewFirebaseAppCheckVerifier builds the SDK client. This blocks on the JWKS fetch.
func NewFirebaseAppCheckVerifier(ctx context.Context, projectID string, guard ReplayGuard) (AppCheckVerifier, error) {
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID})
	if err != nil {
		return nil, fmt.Errorf("init firebase app: %w", err)
	}
	client, err := app.AppCheck(ctx)
	if err != nil {
		return nil, fmt.Errorf("init firebase app check: %w", err)
	}
	return &firebaseAppCheckVerifier{client: client, guard: guard}, nil
}

var (
	errAppCheckTokenMissing = errors.New("app check token is empty")
	errAppCheckNotReady     = errors.New("app check verifier is not ready")
	errAppCheckReplayed     = errors.New("app check token was already used")
)

// Verify fails closed: any problem, including a panic inside the SDK on
// unexpected claims, is reported as an error. After the SDK accepts the token
// it is consumed once through the replay guard, keyed on its jti claim (or a
// SHA-256 of the raw token when jti is absent).
func (v *firebaseAppCheckVerifier) Verify(_ context.Context, token string) (err error) {
	if token == "" {
		return errAppCheckTokenMissing
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("app check verification panicked: %v", r)
		}
	}()

	decoded, verifyErr := v.client.VerifyToken(token)
	if verifyErr != nil {
		return fmt.Errorf("app check token rejected: %w", verifyErr)
	}
	if decoded == nil {
		return errors.New("app check returned no token")
	}

	if !v.guard.Consume(replayKey(token, decoded), decoded.ExpiresAt) {
		return errAppCheckReplayed
	}
	return nil
}

func replayKey(token string, decoded *appcheck.DecodedAppCheckToken) string {
	if jti, ok := decoded.Claims["jti"].(string); ok && jti != "" {
		return "jti:" + jti
	}
	sum := sha256.Sum256([]byte(token))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// lazyAppCheckVerifier answers "not ready" until its background init succeeds.
type lazyAppCheckVerifier struct {
	inner atomic.Pointer[AppCheckVerifier]
}

func newLazyAppCheckVerifier(
	ctx context.Context,
	build func(context.Context) (AppCheckVerifier, error),
	minBackoff, maxBackoff time.Duration,
) *lazyAppCheckVerifier {
	l := &lazyAppCheckVerifier{}
	go l.initLoop(ctx, build, minBackoff, maxBackoff)
	return l
}

func (l *lazyAppCheckVerifier) initLoop(
	ctx context.Context,
	build func(context.Context) (AppCheckVerifier, error),
	minBackoff, maxBackoff time.Duration,
) {
	backoff := minBackoff
	for {
		v, err := build(ctx)
		if err == nil {
			l.inner.Store(&v)
			logger.Infof(ctx, "firebase app check is ready")
			return
		}
		logger.Errorf(ctx, "firebase app check is NOT ready, requests with %s are rejected until it is (retrying in %s): %s", "X-Firebase-AppCheck", backoff, err)

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, maxBackoff)
	}
}

func (l *lazyAppCheckVerifier) Verify(ctx context.Context, token string) error {
	inner := l.inner.Load()
	if inner == nil {
		return errAppCheckNotReady
	}
	return (*inner).Verify(ctx, token)
}
