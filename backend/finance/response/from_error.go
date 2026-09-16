package response

import (
	"errors"

	"sen1or/letslive/finance/domains"

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
	case errors.Is(err, domains.ErrUnauthorized):
		return NewResponseFromTemplate[any](RES_ERR_UNAUTHORIZED, nil, nil, nil)
	case errors.Is(err, domains.ErrDatabaseQuery):
		return NewResponseFromTemplate[any](RES_ERR_DATABASE_QUERY, nil, nil, nil)
	case errors.Is(err, domains.ErrDatabaseIssue):
		return NewResponseFromTemplate[any](RES_ERR_DATABASE_ISSUE, nil, nil, nil)

	case errors.Is(err, domains.ErrAccountNotFound):
		return NewResponseFromTemplate[any](RES_ERR_ACCOUNT_NOT_FOUND, nil, nil, nil)
	case errors.Is(err, domains.ErrAccountFrozen):
		return NewResponseFromTemplate[any](RES_ERR_ACCOUNT_FROZEN, nil, nil, nil)
	case errors.Is(err, domains.ErrInsufficientBalance):
		return NewResponseFromTemplate[any](RES_ERR_INSUFFICIENT_BALANCE, nil, nil, nil)
	case errors.Is(err, domains.ErrInvalidAmount):
		return NewResponseFromTemplate[any](RES_ERR_INVALID_AMOUNT, nil, nil, nil)
	case errors.Is(err, domains.ErrTransactionFailed):
		return NewResponseFromTemplate[any](RES_ERR_TRANSACTION_FAILED, nil, nil, nil)
	case errors.Is(err, domains.ErrTransactionNotFound):
		return NewResponseFromTemplate[any](RES_ERR_TRANSACTION_NOT_FOUND, nil, nil, nil)
	case errors.Is(err, domains.ErrPaymentFailed):
		return NewResponseFromTemplate[any](RES_ERR_PAYMENT_FAILED, nil, nil, nil)
	case errors.Is(err, domains.ErrPaymentNotFound):
		return NewResponseFromTemplate[any](RES_ERR_PAYMENT_NOT_FOUND, nil, nil, nil)
	case errors.Is(err, domains.ErrUnsupportedCurrency):
		return NewResponseFromTemplate[any](RES_ERR_UNSUPPORTED_CURRENCY, nil, nil, nil)
	case errors.Is(err, domains.ErrDepositLimitExceeded):
		return NewResponseFromTemplate[any](RES_ERR_DEPOSIT_LIMIT_EXCEEDED, nil, nil, nil)
	case errors.Is(err, domains.ErrShopItemNotFound):
		return NewResponseFromTemplate[any](RES_ERR_SHOP_ITEM_NOT_FOUND, nil, nil, nil)
	case errors.Is(err, domains.ErrUserService):
		return NewResponseFromTemplate[any](RES_ERR_USER_SERVICE_ERROR, nil, nil, nil)

	default:
		return NewResponseFromTemplate[any](RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	}
}
