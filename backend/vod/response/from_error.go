package response

import (
	"errors"
	"sen1or/letslive/vod/domains"

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
	case errors.Is(err, domains.ErrInvalidPayload):
		return NewResponseFromTemplate[any](RES_ERR_INVALID_PAYLOAD, nil, nil, nil)
	case errors.Is(err, domains.ErrForbidden):
		return NewResponseFromTemplate[any](RES_ERR_FORBIDDEN, nil, nil, nil)
	case errors.Is(err, domains.ErrDatabaseQuery):
		return NewResponseFromTemplate[any](RES_ERR_DATABASE_QUERY, nil, nil, nil)
	case errors.Is(err, domains.ErrDatabaseIssue):
		return NewResponseFromTemplate[any](RES_ERR_DATABASE_ISSUE, nil, nil, nil)

	case errors.Is(err, domains.ErrVODNotFound):
		return NewResponseFromTemplate[any](RES_ERR_VOD_NOT_FOUND, nil, nil, nil)
	case errors.Is(err, domains.ErrVODCreateFailed):
		return NewResponseFromTemplate[any](RES_ERR_VOD_CREATE_FAILED, nil, nil, nil)
	case errors.Is(err, domains.ErrVODUpdateFailed):
		return NewResponseFromTemplate[any](RES_ERR_VOD_UPDATE_FAILED, nil, nil, nil)
	case errors.Is(err, domains.ErrVODViewThreshold):
		return NewResponseFromTemplate[any](RES_ERR_VOD_VIEW_THRESHOLD, nil, nil, nil)

	case errors.Is(err, domains.ErrCommentNotFound):
		return NewResponseFromTemplate[any](RES_ERR_VOD_COMMENT_NOT_FOUND, nil, nil, nil)
	case errors.Is(err, domains.ErrCommentCreateFailed):
		return NewResponseFromTemplate[any](RES_ERR_VOD_COMMENT_CREATE_FAILED, nil, nil, nil)
	case errors.Is(err, domains.ErrCommentDeleteFailed):
		return NewResponseFromTemplate[any](RES_ERR_VOD_COMMENT_DELETE_FAILED, nil, nil, nil)
	case errors.Is(err, domains.ErrCommentAlreadyLiked):
		return NewResponseFromTemplate[any](RES_ERR_VOD_COMMENT_ALREADY_LIKED, nil, nil, nil)
	case errors.Is(err, domains.ErrCommentNotLiked):
		return NewResponseFromTemplate[any](RES_ERR_VOD_COMMENT_NOT_LIKED, nil, nil, nil)

	default:
		return NewResponseFromTemplate[any](RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	}
}
