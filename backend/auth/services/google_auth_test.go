package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"sen1or/letslive/auth/domains"

	"github.com/gofrs/uuid/v5"
)

// existingAuthRepo returns the same auth record for every email lookup.
type existingAuthRepo struct {
	domains.AuthRepository
	auth *domains.Auth
}

func (r existingAuthRepo) GetByEmail(_ context.Context, _ string) (*domains.Auth, error) {
	return r.auth, nil
}

// A token with characters that would break an unescaped query string.
const testIDToken = "header.pay+load/x&aud=evil#frag"

func TestVerifyIDTokenAndGetUser_checksAudience(t *testing.T) {
	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "web-client")
	t.Setenv("GOOGLE_OAUTH_MOBILE_CLIENT_ID", "mobile-client")

	tests := []struct {
		name    string
		aud     string
		wantErr error
	}{
		{"web client audience", "web-client", nil},
		{"mobile client audience", "mobile-client", nil},
		{"foreign audience", "someone-else", domains.ErrUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("id_token"); got != testIDToken {
					t.Errorf("id_token = %q, want %q", got, testIDToken)
				}
				_ = json.NewEncoder(w).Encode(map[string]string{
					"email":          "user@example.com",
					"email_verified": "true",
					"aud":            tt.aud,
				})
			}))
			defer srv.Close()

			userID := uuid.Must(uuid.NewV4())
			svc := GoogleAuthService{
				repo:         existingAuthRepo{auth: &domains.Auth{Email: "user@example.com", UserId: &userID}},
				tokenInfoURL: srv.URL,
			}

			auth, err := svc.VerifyIDTokenAndGetUser(context.Background(), testIDToken)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && (auth == nil || *auth.UserId != userID) {
				t.Fatalf("auth = %+v, want the existing user %s", auth, userID)
			}
		})
	}
}
