package response

import (
	"errors"
	"fmt"
	"testing"

	"sen1or/letslive/finance/domains"
)

// Every domain error must produce exactly the envelope the pre-refactor code
// built at the failure site. A mismatch here is a client-visible API change.
func TestFromErrorMapsToOriginalTemplate(t *testing.T) {
	cases := []struct {
		err      error
		expected ResponseTemplate
	}{
		{domains.ErrInvalidInput, RES_ERR_INVALID_INPUT},
		{domains.ErrUnauthorized, RES_ERR_UNAUTHORIZED},
		{domains.ErrDatabaseQuery, RES_ERR_DATABASE_QUERY},
		{domains.ErrDatabaseIssue, RES_ERR_DATABASE_ISSUE},
		{domains.ErrInternal, RES_ERR_INTERNAL_SERVER},
		{domains.ErrAccountNotFound, RES_ERR_ACCOUNT_NOT_FOUND},
		{domains.ErrAccountFrozen, RES_ERR_ACCOUNT_FROZEN},
		{domains.ErrInsufficientBalance, RES_ERR_INSUFFICIENT_BALANCE},
		{domains.ErrInvalidAmount, RES_ERR_INVALID_AMOUNT},
		{domains.ErrTransactionFailed, RES_ERR_TRANSACTION_FAILED},
		{domains.ErrTransactionNotFound, RES_ERR_TRANSACTION_NOT_FOUND},
		{domains.ErrPaymentFailed, RES_ERR_PAYMENT_FAILED},
		{domains.ErrPaymentNotFound, RES_ERR_PAYMENT_NOT_FOUND},
		{domains.ErrUnsupportedCurrency, RES_ERR_UNSUPPORTED_CURRENCY},
		{domains.ErrDepositLimitExceeded, RES_ERR_DEPOSIT_LIMIT_EXCEEDED},
		{domains.ErrShopItemNotFound, RES_ERR_SHOP_ITEM_NOT_FOUND},
		{domains.ErrUserService, RES_ERR_USER_SERVICE_ERROR},
	}

	for _, tc := range cases {
		t.Run(tc.err.Error(), func(t *testing.T) {
			got := FromError(tc.err)
			if got.StatusCode != tc.expected.StatusCode {
				t.Errorf("status = %d, want %d", got.StatusCode, tc.expected.StatusCode)
			}
			if got.Code != tc.expected.Code {
				t.Errorf("code = %d, want %d", got.Code, tc.expected.Code)
			}
			if got.Key != tc.expected.Key {
				t.Errorf("key = %q, want %q", got.Key, tc.expected.Key)
			}
			if got.Success {
				t.Error("success = true, want false")
			}
		})
	}
}

// The wallet, transaction and deposit services branch on "is this
// account-not-found?" — that check used to compare an HTTP business code and
// now uses errors.Is, so wrapping must stay transparent.
func TestFromErrorSeesThroughWrapping(t *testing.T) {
	wrapped := fmt.Errorf("getuserwalletbyownerid: %w", domains.ErrAccountNotFound)

	if !errors.Is(wrapped, domains.ErrAccountNotFound) {
		t.Fatal("errors.Is must see through the wrapping, or the services misroute")
	}
	if got := FromError(wrapped); got.Key != RES_ERR_ACCOUNT_NOT_FOUND_KEY {
		t.Errorf("key = %q, want %q", got.Key, RES_ERR_ACCOUNT_NOT_FOUND_KEY)
	}
}

func TestFromErrorNil(t *testing.T) {
	if got := FromError(nil); got != nil {
		t.Errorf("FromError(nil) = %+v, want nil", got)
	}
}

// An unrecognised error must not leak its text to the client.
func TestFromErrorUnknownIsInternal(t *testing.T) {
	got := FromError(errors.New("stripe: invalid api key provided sk_live_xxx"))
	if got.Key != RES_ERR_INTERNAL_SERVER_KEY {
		t.Errorf("key = %q, want %q", got.Key, RES_ERR_INTERNAL_SERVER_KEY)
	}
	if got.Message != RES_ERR_INTERNAL_SERVER.Message {
		t.Errorf("message = %q, want the generic internal message", got.Message)
	}
}
