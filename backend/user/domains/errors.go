package domains

import "errors"

// Domain errors returned by repositories and services.
//
// Nothing below the handler decides an HTTP status, a business code or an
// i18n key: those belong to the transport layer, which maps these values in
// response.FromError. Wrap them with fmt.Errorf("%w") to add context; the
// mapping uses errors.Is.
var (
	ErrInvalidInput  = errors.New("invalid input")
	ErrDatabaseQuery = errors.New("database query failed")
	ErrDatabaseIssue = errors.New("database issue")
	ErrInternal      = errors.New("internal error")

	ErrUserNotFound          = errors.New("user not found")
	ErrUsernameTaken         = errors.New("username already taken")
	ErrNotificationNotFound  = errors.New("notification not found")
	ErrInsufficientInventory = errors.New("insufficient inventory")
)
