package serviceresponse

import (
	"errors"

	"sen1or/letslive/auth/domains"

	"github.com/go-playground/validator/v10"
)

// FromError maps a domain error to the HTTP envelope.
//
// This is the single place where a failure below the handler acquires a
// status code, a business code and an i18n key. An unrecognised error is
// deliberately reported as a generic internal error rather than leaking its
// text to the client — which matters most here, where the underlying errors
// come from password hashing, OAuth exchanges and token parsing.
func FromError(err error) *Response[any] {
	if err == nil {
		return nil
	}

	// a failure from another service keeps its own status, code and key
	var downstream *domains.DownstreamError
	if errors.As(err, &downstream) {
		return NewResponse[any](false, downstream.StatusCode, downstream.Code, downstream.Key, downstream.Message, nil, nil, nil)
	}

	// validation failures carry per-field details, so they are formatted
	// before the sentinel switch; errors.As sees through the %w wrapping
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		return NewResponseWithValidationErrors[any](nil, nil, validationErrs)
	}

	switch {
	case errors.Is(err, domains.ErrInvalidInput):
		return NewResponseFromTemplate[any](RES_ERR_INVALID_INPUT, nil, nil, nil)
	case errors.Is(err, domains.ErrUnauthorized):
		return NewResponseFromTemplate[any](RES_ERR_UNAUTHORIZED, nil, nil, nil)
	case errors.Is(err, domains.ErrDatabaseQuery):
		return NewResponseFromTemplate[any](RES_ERR_DATABASE_QUERY, nil, nil, nil)
	case errors.Is(err, domains.ErrDatabaseIssue):
		return NewResponseFromTemplate[any](RES_ERR_DATABASE_ISSUE, nil, nil, nil)

	case errors.Is(err, domains.ErrAuthNotFound):
		return NewResponseFromTemplate[any](RES_ERR_AUTH_NOT_FOUND, nil, nil, nil)
	case errors.Is(err, domains.ErrAuthAlreadyExists):
		return NewResponseFromTemplate[any](RES_ERR_AUTH_ALREADY_EXISTS, nil, nil, nil)
	case errors.Is(err, domains.ErrEmailOrPasswordIncorrect):
		return NewResponseFromTemplate[any](RES_ERR_EMAIL_OR_PASSWORD_INCORRECT, nil, nil, nil)
	case errors.Is(err, domains.ErrPasswordNotMatch):
		return NewResponseFromTemplate[any](RES_ERR_PASSWORD_NOT_MATCH, nil, nil, nil)
	case errors.Is(err, domains.ErrCaptchaFailed):
		return NewResponseFromTemplate[any](RES_ERR_CAPTCHA_FAILED, nil, nil, nil)
	case errors.Is(err, domains.ErrRefreshTokenNotFound):
		return NewResponseFromTemplate[any](RES_ERR_REFRESH_TOKEN_NOT_FOUND, nil, nil, nil)
	case errors.Is(err, domains.ErrSignUpOTPNotFound):
		return NewResponseFromTemplate[any](RES_ERR_SIGN_UP_OTP_NOT_FOUND, nil, nil, nil)
	case errors.Is(err, domains.ErrSignUpOTPExpired):
		return NewResponseFromTemplate[any](RES_ERR_SIGN_UP_OTP_EXPIRED, nil, nil, nil)
	case errors.Is(err, domains.ErrSignUpOTPAlreadyUsed):
		return NewResponseFromTemplate[any](RES_ERR_SIGN_UP_OTP_ALREADY_USED, nil, nil, nil)
	case errors.Is(err, domains.ErrFailedToCreateSignUpOTP):
		return NewResponseFromTemplate[any](RES_ERR_FAILED_TO_CREATE_SIGN_UP_OTP, nil, nil, nil)
	case errors.Is(err, domains.ErrFailedToSendVerification):
		return NewResponseFromTemplate[any](RES_ERR_FAILED_TO_SEND_VERIFICATION, nil, nil, nil)

	default:
		return NewResponseFromTemplate[any](RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	}
}
