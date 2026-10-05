package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestOAuthGoogleLoginRemembersSafeRedirect(t *testing.T) {
	h := &AuthHandler{}
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/google?redirectUrl="+url.QueryEscape("/messages?c=1"), nil)
	rec := httptest.NewRecorder()

	h.OAuthGoogleLoginHandler(rec, req)

	c := findCookie(rec.Result().Cookies(), oauthRedirectCookieName)
	if c == nil || c.MaxAge < 0 {
		t.Fatalf("expected %s cookie to be set, got %+v", oauthRedirectCookieName, c)
	}
	got, err := url.QueryUnescape(c.Value)
	if err != nil || got != "/messages?c=1" {
		t.Fatalf("cookie value = %q (%v), want /messages?c=1", c.Value, err)
	}
	if !c.HttpOnly || !c.Secure {
		t.Fatalf("cookie must be HttpOnly and Secure: %+v", c)
	}
}

func TestOAuthGoogleLoginClearsRedirectWhenNoneGiven(t *testing.T) {
	h := &AuthHandler{}
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/google", nil)
	rec := httptest.NewRecorder()

	h.OAuthGoogleLoginHandler(rec, req)

	c := findCookie(rec.Result().Cookies(), oauthRedirectCookieName)
	if c == nil || c.MaxAge >= 0 {
		t.Fatalf("expected %s cookie to be cleared, got %+v", oauthRedirectCookieName, c)
	}
}

func TestOAuthSuccessRedirectURL(t *testing.T) {
	t.Setenv("CLIENT_URL", "https://letslive.work")

	cases := []struct {
		name   string
		cookie *http.Cookie
		want   string
	}{
		{"no cookie", nil, "https://letslive.work/login"},
		{"remembered path", &http.Cookie{Name: oauthRedirectCookieName, Value: url.QueryEscape("/messages")}, "https://letslive.work/login?redirectUrl=%2Fmessages"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/auth/google/callback", nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}
			if got := oauthSuccessRedirectURL(req); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
