package domains

import "errors"

// Domain errors returned by repositories and services.
//
// Nothing below the handler decides an HTTP status, a business code or an
// i18n key: those belong to the transport layer, which maps these values in
// response.FromError. Wrap them with fmt.Errorf("%w") to add context; the
// mapping uses errors.Is.
var (
	ErrInvalidPayload = errors.New("invalid payload")
	ErrForbidden      = errors.New("forbidden")
	ErrDatabaseQuery  = errors.New("database query failed")
	ErrDatabaseIssue  = errors.New("database issue")
	ErrInternal       = errors.New("internal error")

	ErrLivestreamNotFound         = errors.New("livestream not found")
	ErrLivestreamCreateFailed     = errors.New("failed to create livestream record")
	ErrLivestreamUpdateFailed     = errors.New("failed to update livestream record")
	ErrLivestreamUpdateAfterEnded = errors.New("cannot update an ended livestream")
	ErrEndAlreadyEndedLivestream  = errors.New("livestream already ended")
	ErrVODCreateFailed            = errors.New("failed to create vod record")
)
