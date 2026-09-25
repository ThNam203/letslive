package response

import (
	"errors"

	"sen1or/letslive/user/domains"

	"github.com/go-playground/validator/v10"
)

// FromError maps a domain error to the HTTP envelope.
//
// This is the single place where a failure below the handler acquires a
// status code, a business code and an i18n key. An unrecognised error is
// deliberately reported as a generic internal error rather than leaking its
// text to the client.
func FromError(err error) *Response[any] {
	if err == nil {
		return nil
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
	case errors.Is(err, domains.ErrDatabaseQuery):
		return NewResponseFromTemplate[any](RES_ERR_DATABASE_QUERY, nil, nil, nil)
	case errors.Is(err, domains.ErrDatabaseIssue):
		return NewResponseFromTemplate[any](RES_ERR_DATABASE_ISSUE, nil, nil, nil)

	case errors.Is(err, domains.ErrUserNotFound):
		return NewResponseFromTemplate[any](RES_ERR_USER_NOT_FOUND, nil, nil, nil)
	case errors.Is(err, domains.ErrUsernameTaken):
		return NewResponseFromTemplate[any](RES_ERR_USERNAME_TAKEN, nil, nil, nil)
	case errors.Is(err, domains.ErrNotificationNotFound):
		return NewResponseFromTemplate[any](RES_ERR_NOTIFICATION_NOT_FOUND, nil, nil, nil)
	case errors.Is(err, domains.ErrAccountDisabled):
		return NewResponseFromTemplate[any](RES_ERR_ACCOUNT_DISABLED, nil, nil, nil)

	default:
		return NewResponseFromTemplate[any](RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	}
}
