package utils

import (
	"context"
	"errors"
	"testing"

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

func TestFirebaseAppCheckVerifier_Verify(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		inner   *fakeTokenVerifier
		wantErr bool
		wantSee bool
	}{
		{"valid token", "t", &fakeTokenVerifier{token: &appcheck.DecodedAppCheckToken{Subject: "app"}}, false, true},
		{"empty token never reaches the sdk", "", &fakeTokenVerifier{token: &appcheck.DecodedAppCheckToken{}}, true, false},
		{"sdk rejection is an error", "t", &fakeTokenVerifier{err: errors.New("expired")}, true, true},
		{"nil decoded token is an error", "t", &fakeTokenVerifier{}, true, true},
		{"sdk panic fails closed", "t", &fakeTokenVerifier{boom: true}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &firebaseAppCheckVerifier{client: tt.inner}
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

func TestNewAppCheckVerifierFromEnv_disabledWithoutProjectID(t *testing.T) {
	t.Setenv(FirebaseProjectIDEnv, "")
	v, err := NewAppCheckVerifierFromEnv(context.Background())
	if err != nil {
		t.Fatalf("unset project id must not be an error: %v", err)
	}
	if v != nil {
		t.Fatal("expected no verifier (App Check off) when FIREBASE_PROJECT_ID is unset")
	}
}
