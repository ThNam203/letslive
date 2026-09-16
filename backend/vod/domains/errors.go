package domains

import "errors"

// Domain errors returned by repositories and services.
//
// Nothing below the handler decides an HTTP status, a business code or an
// i18n key: those belong to the transport layer, which maps these values in
// response.FromError. Wrap them with fmt.Errorf("%w") to add context; the
// mapping uses errors.Is.
var (
	ErrInvalidInput   = errors.New("invalid input")
	ErrInvalidPayload = errors.New("invalid payload")
	ErrForbidden      = errors.New("forbidden")
	ErrDatabaseQuery  = errors.New("database query failed")
	ErrDatabaseIssue  = errors.New("database issue")
	ErrInternal       = errors.New("internal error")

	ErrVODNotFound      = errors.New("vod not found")
	ErrVODCreateFailed  = errors.New("failed to create vod record")
	ErrVODUpdateFailed  = errors.New("failed to update vod record")
	ErrVODViewThreshold = errors.New("watch time threshold not met")

	ErrCommentNotFound     = errors.New("comment not found")
	ErrCommentCreateFailed = errors.New("failed to create comment")
	ErrCommentDeleteFailed = errors.New("failed to delete comment")
	ErrCommentAlreadyLiked = errors.New("comment already liked")
	ErrCommentNotLiked     = errors.New("comment has not been liked")
)
