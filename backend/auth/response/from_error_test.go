package serviceresponse

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"sen1or/letslive/auth/domains"
)

// Every domain error must produce exactly the envelope the pre-refactor code
// built at the failure site. A mismatch here is a client-visible API change.
func TestFromErrorMapsToOriginalTemplate(t *testing.T) {
	cases := []struct {
		err      error
		expected ResponseTemplate
	}{
		{domains.ErrInvalidInput, RES_ERR_INVALID_INPUT},
		{domains.ErrUnauthorized, RES_ERR_UNAUTHORIZED},
		{domains.ErrDatabaseQuery, RES_ERR_DATABASE_QUERY},
		{domains.ErrDatabaseIssue, RES_ERR_DATABASE_ISSUE},
		{domains.ErrInternal, RES_ERR_INTERNAL_SERVER},
		{domains.ErrAuthNotFound, RES_ERR_AUTH_NOT_FOUND},
		{domains.ErrAuthAlreadyExists, RES_ERR_AUTH_ALREADY_EXISTS},
		{domains.ErrEmailOrPasswordIncorrect, RES_ERR_EMAIL_OR_PASSWORD_INCORRECT},
		{domains.ErrPasswordNotMatch, RES_ERR_PASSWORD_NOT_MATCH},
		{domains.ErrCaptchaFailed, RES_ERR_CAPTCHA_FAILED},
		{domains.ErrRefreshTokenNotFound, RES_ERR_REFRESH_TOKEN_NOT_FOUND},
		{domains.ErrSignUpOTPNotFound, RES_ERR_SIGN_UP_OTP_NOT_FOUND},
		{domains.ErrSignUpOTPExpired, RES_ERR_SIGN_UP_OTP_EXPIRED},
		{domains.ErrSignUpOTPAlreadyUsed, RES_ERR_SIGN_UP_OTP_ALREADY_USED},
		{domains.ErrFailedToCreateSignUpOTP, RES_ERR_FAILED_TO_CREATE_SIGN_UP_OTP},
		{domains.ErrFailedToSendVerification, RES_ERR_FAILED_TO_SEND_VERIFICATION},
	}

	for _, tc := range cases {
		t.Run(tc.err.Error(), func(t *testing.T) {
			got := FromError(tc.err)
			if got.StatusCode != tc.expected.StatusCode {
				t.Errorf("status = %d, want %d", got.StatusCode, tc.expected.StatusCode)
			}
			if got.Code != tc.expected.Code {
				t.Errorf("code = %d, want %d", got.Code, tc.expected.Code)
			}
			if got.Key != tc.expected.Key {
				t.Errorf("key = %q, want %q", got.Key, tc.expected.Key)
			}
			if got.Success {
				t.Error("success = true, want false")
			}
		})
	}
}

// Sign-up calls the user service; its errors must reach the client with the
// downstream code and key intact, or a taken username reads as a server fault.
func TestFromErrorForwardsDownstreamError(t *testing.T) {
	got := FromError(&domains.DownstreamError{
		StatusCode: http.StatusConflict,
		Code:       30003,
		Key:        "res_err_username_taken",
		Message:    "Username already taken.",
	})

	if got.StatusCode != http.StatusConflict {
		t.Errorf("status = %d, want %d", got.StatusCode, http.StatusConflict)
	}
	if got.Code != 30003 {
		t.Errorf("code = %d, want 30003", got.Code)
	}
	if got.Key != "res_err_username_taken" {
		t.Errorf("key = %q, want res_err_username_taken", got.Key)
	}
	if got.Success {
		t.Error("success = true, want false")
	}
}

// Services wrap sentinels (validation failures pair ErrInvalidInput with the
// validator error), so the mapping has to see through fmt.Errorf.
func TestFromErrorSeesThroughWrapping(t *testing.T) {
	wrapped := fmt.Errorf("email=%v: %w", "a@b.com", domains.ErrAuthNotFound)

	if got := FromError(wrapped); got.Key != RES_ERR_AUTH_NOT_FOUND_KEY {
		t.Errorf("key = %q, want %q", got.Key, RES_ERR_AUTH_NOT_FOUND_KEY)
	}
}

// If a caller ever does wrap a sentinel with request context, that context must
// stay server-side: the client gets the template message, never the wrapped
// identifiers. Repositories log such values instead of carrying them.
func TestFromErrorDoesNotLeakWrappedContext(t *testing.T) {
	wrapped := fmt.Errorf("code=%v email=%v: %w", "483920", "a@b.com", domains.ErrSignUpOTPNotFound)

	got := FromError(wrapped)
	if got.Message != RES_ERR_SIGN_UP_OTP_NOT_FOUND.Message {
		t.Errorf("message = %q, want the template message", got.Message)
	}
	if got.ErrorDetails != nil {
		t.Errorf("errorDetails = %v, want nil", got.ErrorDetails)
	}
}

func TestFromErrorNil(t *testing.T) {
	if got := FromError(nil); got != nil {
		t.Errorf("FromError(nil) = %+v, want nil", got)
	}
}

// Auth failures wrap bcrypt, OAuth and token-parsing errors; none of that text
// may reach the client.
func TestFromErrorUnknownIsInternal(t *testing.T) {
	got := FromError(errors.New("oauth2: cannot fetch token: 401 Unauthorized, client_secret=abc123"))

	if got.Key != RES_ERR_INTERNAL_SERVER_KEY {
		t.Errorf("key = %q, want %q", got.Key, RES_ERR_INTERNAL_SERVER_KEY)
	}
	if got.Message != RES_ERR_INTERNAL_SERVER.Message {
		t.Errorf("message = %q, want the generic internal message", got.Message)
	}
}
