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
	ErrUnauthorized  = errors.New("unauthorized")
	ErrDatabaseQuery = errors.New("database query failed")
	ErrDatabaseIssue = errors.New("database issue")
	ErrInternal      = errors.New("internal error")

	ErrAccountNotFound      = errors.New("account not found")
	ErrAccountFrozen        = errors.New("account frozen")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrTransactionFailed    = errors.New("transaction failed")
	ErrTransactionNotFound  = errors.New("transaction not found")
	ErrPaymentFailed        = errors.New("payment failed")
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrUnsupportedCurrency  = errors.New("unsupported currency")
	ErrDepositLimitExceeded = errors.New("deposit limit exceeded")
	ErrShopItemNotFound     = errors.New("shop item not found")
	ErrUserService          = errors.New("user service error")
)
