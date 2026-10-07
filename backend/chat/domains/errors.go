package domains

import "errors"

// Domain errors returned by repositories and services. The transport layer
// maps them to status, code and i18n key in response.FromError.
var (
	ErrInvalidInput  = errors.New("invalid input")
	ErrForbidden     = errors.New("forbidden")
	ErrDatabaseIssue = errors.New("database issue")
	ErrUserService   = errors.New("user service error")
	ErrAlreadyExists = errors.New("already exists")

	ErrConversationNotFound = errors.New("conversation not found")
	ErrNotParticipant       = errors.New("not a participant")
	ErrInsufficientRole     = errors.New("insufficient role")
	ErrDmMessageNotFound    = errors.New("dm message not found")
	ErrCannotMessageSelf    = errors.New("cannot message self")
	ErrTooManyParticipants  = errors.New("too many participants")
	ErrUserSetupIncomplete  = errors.New("user setup incomplete")
	ErrRoomNotFound         = errors.New("room not found")
)
