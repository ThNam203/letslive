package response

import (
	"errors"

	"sen1or/letslive/livestream/domains"

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
	case errors.Is(err, domains.ErrInvalidPayload):
		return NewResponseFromTemplate[any](RES_ERR_INVALID_PAYLOAD, nil, nil, nil)
	case errors.Is(err, domains.ErrForbidden):
		return NewResponseFromTemplate[any](RES_ERR_FORBIDDEN, nil, nil, nil)
	case errors.Is(err, domains.ErrDatabaseQuery):
		return NewResponseFromTemplate[any](RES_ERR_DATABASE_QUERY, nil, nil, nil)
	case errors.Is(err, domains.ErrDatabaseIssue):
		return NewResponseFromTemplate[any](RES_ERR_DATABASE_ISSUE, nil, nil, nil)

	case errors.Is(err, domains.ErrLivestreamNotFound):
		return NewResponseFromTemplate[any](RES_ERR_LIVESTREAM_NOT_FOUND, nil, nil, nil)
	case errors.Is(err, domains.ErrLivestreamCreateFailed):
		return NewResponseFromTemplate[any](RES_ERR_LIVESTREAM_CREATE_FAILED, nil, nil, nil)
	case errors.Is(err, domains.ErrLivestreamUpdateFailed):
		return NewResponseFromTemplate[any](RES_ERR_LIVESTREAM_UPDATE_FAILED, nil, nil, nil)
	case errors.Is(err, domains.ErrLivestreamUpdateAfterEnded):
		return NewResponseFromTemplate[any](RES_ERR_LIVESTREAM_UPDATE_AFTER_ENDED, nil, nil, nil)
	case errors.Is(err, domains.ErrEndAlreadyEndedLivestream):
		return NewResponseFromTemplate[any](RES_ERR_END_ALREADY_ENDED_LIVESTREAM, nil, nil, nil)
	case errors.Is(err, domains.ErrVODCreateFailed):
		return NewResponseFromTemplate[any](RES_ERR_VOD_CREATE_FAILED, nil, nil, nil)

	default:
		return NewResponseFromTemplate[any](RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	}
}
