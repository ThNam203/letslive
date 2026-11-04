package utils

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"firebase.google.com/go/v4/appcheck"
)

type fakeTokenVerifier struct {
	token *appcheck.DecodedAppCheckToken
	err   error
	boom  bool
	seen  string
}

func (f *fakeTokenVerifier) VerifyToken(token string) (*appcheck.DecodedAppCheckToken, error) {
	f.seen = token
	if f.boom {
		panic("malformed claims")
	}
	return f.token, f.err
}

func decoded(jti string) *appcheck.DecodedAppCheckToken {
	claims := map[string]interface{}{}
	if jti != "" {
		claims["jti"] = jti
	}
	return &appcheck.DecodedAppCheckToken{Subject: "app", ExpiresAt: time.Now().Add(time.Hour), Claims: claims}
}

func TestFirebaseAppCheckVerifier_Verify(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		inner   *fakeTokenVerifier
		wantErr bool
		wantSee bool
	}{
		{"valid token", "t", &fakeTokenVerifier{token: decoded("j1")}, false, true},
		{"valid token without jti", "t", &fakeTokenVerifier{token: decoded("")}, false, true},
		{"empty token never reaches the sdk", "", &fakeTokenVerifier{token: decoded("j1")}, true, false},
		{"sdk rejection is an error", "t", &fakeTokenVerifier{err: errors.New("expired")}, true, true},
		{"nil decoded token is an error", "t", &fakeTokenVerifier{}, true, true},
		{"sdk panic fails closed", "t", &fakeTokenVerifier{boom: true}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &firebaseAppCheckVerifier{client: tt.inner, guard: NewMemoryReplayGuard(10)}
			err := v.Verify(context.Background(), tt.token)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if (tt.inner.seen != "") != tt.wantSee {
				t.Fatalf("sdk saw %q, wantSee %v", tt.inner.seen, tt.wantSee)
			}
		})
	}
}

func TestFirebaseAppCheckVerifier_replayedTokenIsRejected(t *testing.T) {
	tests := []struct {
		name string
		jti  string
	}{
		{"keyed on jti", "jti-1"},
		{"keyed on the token hash when jti is absent", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &firebaseAppCheckVerifier{client: &fakeTokenVerifier{token: decoded(tt.jti)}, guard: NewMemoryReplayGuard(10)}
			if err := v.Verify(context.Background(), "tok"); err != nil {
				t.Fatalf("first use: %v", err)
			}
			if err := v.Verify(context.Background(), "tok"); !errors.Is(err, errAppCheckReplayed) {
				t.Fatalf("second use err = %v, want replay rejection", err)
			}
		})
	}
}

func TestNewAppCheckVerifierFromEnv(t *testing.T) {
	cancelled, cancel := context.WithCancel(context.Background())
	cancel() // the background init must not touch the network in a test

	tests := []struct {
		name    string
		env     string
		wantNil bool
		wantErr bool
	}{
		{"unset", "", true, false},
		{"blank is trimmed to off", "  \t ", true, false},
		{"malformed project id fails hard", "not a project!", false, true},
		{"too short fails hard", "abc", false, true},
		{"padded valid id is accepted", "  my-project-123 ", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(FirebaseProjectIDEnv, tt.env)
			v, err := NewAppCheckVerifierFromEnv(cancelled)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if (v == nil) != (tt.wantNil || tt.wantErr) {
				t.Fatalf("verifier nil = %v", v == nil)
			}
		})
	}
}

func TestLazyAppCheckVerifier_notReadyRejectsThenReadyAccepts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	release := make(chan struct{})
	ready := &firebaseAppCheckVerifier{client: &fakeTokenVerifier{token: decoded("j")}, guard: NewMemoryReplayGuard(10)}
	build := func(context.Context) (AppCheckVerifier, error) {
		<-release
		return ready, nil
	}
	l := newLazyAppCheckVerifier(ctx, build, time.Millisecond, time.Millisecond)

	if err := l.Verify(ctx, "tok"); !errors.Is(err, errAppCheckNotReady) {
		t.Fatalf("before ready err = %v, want not-ready", err)
	}
	close(release)
	waitFor(t, func() bool { return l.inner.Load() != nil })
	if err := l.Verify(ctx, "tok"); err != nil {
		t.Fatalf("after ready: %v", err)
	}
}

func TestLazyAppCheckVerifier_initFailureRetriesWithoutCrashing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var attempts atomic.Int32
	ready := &firebaseAppCheckVerifier{client: &fakeTokenVerifier{token: decoded("j")}, guard: NewMemoryReplayGuard(10)}
	build := func(context.Context) (AppCheckVerifier, error) {
		if attempts.Add(1) < 3 {
			return nil, errors.New("jwks unreachable")
		}
		return ready, nil
	}
	l := newLazyAppCheckVerifier(ctx, build, time.Millisecond, 2*time.Millisecond)

	waitFor(t, func() bool { return l.inner.Load() != nil })
	if got := attempts.Load(); got != 3 {
		t.Fatalf("attempts = %d, want 3", got)
	}
	if err := l.Verify(ctx, "tok"); err != nil {
		t.Fatalf("after retry: %v", err)
	}
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for condition")
		}
		time.Sleep(time.Millisecond)
	}
}
