package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/services"

	"github.com/gofrs/uuid/v5"
)

const captchaFailedKey = "res_err_captcha_failed"

// fakeAuthRepo reports every email as registered with an empty password hash:
// login then ends in "email or password incorrect" and signup verification in
// "already exists", so neither sends mail, and neither is a captcha error.
type fakeAuthRepo struct{ lookups int }

func (f *fakeAuthRepo) GetByID(context.Context, uuid.UUID) (*domains.Auth, error) {
	return nil, domains.ErrAuthNotFound
}
func (f *fakeAuthRepo) GetByUserID(context.Context, uuid.UUID) (*domains.Auth, error) {
	return nil, domains.ErrAuthNotFound
}
func (f *fakeAuthRepo) GetByEmail(context.Context, string) (*domains.Auth, error) {
	f.lookups++
	return &domains.Auth{}, nil
}
func (f *fakeAuthRepo) Create(context.Context, domains.Auth) (*domains.Auth, error) {
	return nil, domains.ErrInternal
}
func (f *fakeAuthRepo) UpdatePasswordHash(context.Context, string, string) error { return nil }

type fakeAppCheck struct {
	err   error
	calls int
	got   string
}

func (f *fakeAppCheck) Verify(_ context.Context, token string) error {
	f.calls++
	f.got = token
	return f.err
}

type fakeTurnstile struct {
	err   error
	calls int
	got   string
}

func (f *fakeTurnstile) check(token, _ string) error {
	f.calls++
	f.got = token
	return f.err
}

type captchaCase struct {
	name          string
	userAgent     string
	appCheck      *fakeAppCheck // nil: App Check is not configured
	appCheckToken string
	turnstileErr  error
	turnstile     string
	wantCaptchaOK bool
	wantAppCalls  int
	wantTurnCalls int
}

func captchaCases() []captchaCase {
	return []captchaCase{
		{
			name: "login with User-Agent dart and no captcha is rejected", userAgent: "dart:io/3.5",
			turnstileErr: domains.ErrCaptchaFailed, wantTurnCalls: 1,
		},
		{
			name: "every formerly whitelisted User-Agent gets no special treatment", userAgent: "okhttp flutter letslive-mobile",
			turnstileErr: domains.ErrCaptchaFailed, wantTurnCalls: 1,
		},
		{
			name: "web login still requires turnstile", userAgent: "Mozilla/5.0",
			turnstileErr: domains.ErrCaptchaFailed, wantTurnCalls: 1,
		},
		{
			name: "web login with a valid turnstile token passes", userAgent: "Mozilla/5.0",
			turnstile: "ok", wantCaptchaOK: true, wantTurnCalls: 1,
		},
		{
			name: "login with valid app check token skips turnstile", userAgent: "dart:io/3.5",
			appCheck: &fakeAppCheck{}, appCheckToken: "good", wantCaptchaOK: true, wantAppCalls: 1,
		},
		{
			name: "login with invalid app check token is rejected", userAgent: "dart:io/3.5",
			appCheck: &fakeAppCheck{err: errors.New("bad signature")}, appCheckToken: "bad",
			wantAppCalls: 1,
		},
		{
			name: "invalid app check token is rejected even with a valid turnstile token", userAgent: "dart:io/3.5",
			appCheck: &fakeAppCheck{err: errors.New("expired")}, appCheckToken: "bad", turnstile: "ok",
			wantAppCalls: 1,
		},
		{
			name: "valid app check token wins over a failing turnstile token", userAgent: "dart:io/3.5",
			appCheck: &fakeAppCheck{}, appCheckToken: "good", turnstile: "garbage",
			turnstileErr: domains.ErrCaptchaFailed, wantCaptchaOK: true, wantAppCalls: 1,
		},
		{
			name: "app check configured but header absent falls back to turnstile", userAgent: "Mozilla/5.0",
			appCheck: &fakeAppCheck{}, turnstile: "ok", wantCaptchaOK: true, wantTurnCalls: 1,
		},
		{
			name: "app check configured, header absent, bad turnstile is rejected", userAgent: "Mozilla/5.0",
			appCheck: &fakeAppCheck{}, turnstileErr: domains.ErrCaptchaFailed, wantTurnCalls: 1,
		},
		{
			name: "app check not configured ignores the header and requires turnstile", userAgent: "dart:io/3.5",
			appCheckToken: "good", turnstileErr: domains.ErrCaptchaFailed, wantTurnCalls: 1,
		},
	}
}

func newCaptchaHandler(c captchaCase, repo *fakeAuthRepo, turn *fakeTurnstile) *AuthHandler {
	h := &AuthHandler{
		authService:  *services.NewAuthService(repo, nil),
		checkCaptcha: turn.check,
	}
	if c.appCheck != nil {
		h.appCheck = c.appCheck
	}
	return h
}

func runCaptchaCase(t *testing.T, c captchaCase, path string, body map[string]string, call func(*AuthHandler, http.ResponseWriter, *http.Request)) {
	t.Helper()
	repo := &fakeAuthRepo{}
	turn := &fakeTurnstile{err: c.turnstileErr}
	h := newCaptchaHandler(c, repo, turn)

	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(payload))
	req.Header.Set("User-Agent", c.userAgent)
	if c.appCheckToken != "" {
		req.Header.Set("X-Firebase-AppCheck", c.appCheckToken)
	}
	rec := httptest.NewRecorder()
	call(h, rec, req)

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("response is not json: %v (%s)", err, rec.Body.String())
	}
	gotCaptchaErr := res["key"] == captchaFailedKey
	if c.wantCaptchaOK == gotCaptchaErr {
		t.Fatalf("captcha passed=%v but response was %d %v", !gotCaptchaErr, rec.Code, res)
	}
	if gotCaptchaErr && rec.Code != http.StatusBadRequest {
		t.Fatalf("captcha failure must be 400, got %d", rec.Code)
	}
	if !c.wantCaptchaOK && repo.lookups != 0 {
		t.Fatal("credentials were checked although the captcha was rejected")
	}
	appCalls := 0
	if c.appCheck != nil {
		appCalls = c.appCheck.calls
		if c.appCheckToken != "" && c.appCheck.got != c.appCheckToken {
			t.Fatalf("verifier saw %q, want the header value", c.appCheck.got)
		}
	}
	if appCalls != c.wantAppCalls {
		t.Fatalf("app check calls = %d, want %d", appCalls, c.wantAppCalls)
	}
	if turn.calls != c.wantTurnCalls {
		t.Fatalf("turnstile calls = %d, want %d", turn.calls, c.wantTurnCalls)
	}
	if turn.calls > 0 && turn.got != c.turnstile {
		t.Fatalf("turnstile saw token %q, want %q", turn.got, c.turnstile)
	}
}

func TestLogInHandler_botCheck(t *testing.T) {
	for _, c := range captchaCases() {
		t.Run(c.name, func(t *testing.T) {
			body := map[string]string{"email": "a@example.com", "password": "Passw0rd!x", "turnstileToken": c.turnstile}
			runCaptchaCase(t, c, "/v1/auth/login", body, (*AuthHandler).LogInHandler)
		})
	}
}

func TestRequestEmailVerificationHandler_botCheck(t *testing.T) {
	for _, c := range captchaCases() {
		t.Run(c.name, func(t *testing.T) {
			body := map[string]string{"email": "a@example.com", "turnstileToken": c.turnstile}
			runCaptchaCase(t, c, "/v1/auth/signup/verify", body, (*AuthHandler).RequestEmailVerificationHandler)
		})
	}
}
