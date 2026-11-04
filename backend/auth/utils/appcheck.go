package utils

import (
	"context"
	"errors"
	"fmt"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/appcheck"
)

// FirebaseProjectIDEnv turns Firebase App Check verification on. App Check
// tokens are verified against Google's public JWKS, so the project id is the
// only configuration needed (no service account).
const FirebaseProjectIDEnv = "FIREBASE_PROJECT_ID"

// AppCheckVerifier verifies a Firebase App Check token. A nil error means the
// token is genuine, unexpired and issued for this project.
type AppCheckVerifier interface {
	Verify(ctx context.Context, token string) error
}

// tokenVerifier is the part of *appcheck.Client we use; it lets tests avoid the network.
type tokenVerifier interface {
	VerifyToken(token string) (*appcheck.DecodedAppCheckToken, error)
}

type firebaseAppCheckVerifier struct {
	client tokenVerifier
}

// NewAppCheckVerifierFromEnv returns (nil, nil) when FIREBASE_PROJECT_ID is
// unset: App Check is then off. When it is set but the SDK cannot start (for
// example the JWKS cannot be fetched) it returns an error and the caller must
// not continue as if App Check were off.
func NewAppCheckVerifierFromEnv(ctx context.Context) (AppCheckVerifier, error) {
	projectID := os.Getenv(FirebaseProjectIDEnv)
	if projectID == "" {
		return nil, nil
	}
	return NewFirebaseAppCheckVerifier(ctx, projectID)
}

func NewFirebaseAppCheckVerifier(ctx context.Context, projectID string) (AppCheckVerifier, error) {
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID})
	if err != nil {
		return nil, fmt.Errorf("init firebase app: %w", err)
	}
	client, err := app.AppCheck(ctx)
	if err != nil {
		return nil, fmt.Errorf("init firebase app check: %w", err)
	}
	return &firebaseAppCheckVerifier{client: client}, nil
}

var errAppCheckTokenMissing = errors.New("app check token is empty")

// Verify fails closed: any problem, including a panic inside the SDK on
// unexpected claims, is reported as an error.
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
	return nil
}
