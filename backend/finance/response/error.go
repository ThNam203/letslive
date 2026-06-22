package response

import "net/http"

// Error codes
const (
	// Generic (shared range 20000-20017)
	RES_ERR_INVALID_INPUT_CODE   = 20000
	RES_ERR_INVALID_PAYLOAD_CODE = 20001
	RES_ERR_UNAUTHORIZED_CODE    = 20005
	RES_ERR_FORBIDDEN_CODE       = 20008
	RES_ERR_ROUTE_NOT_FOUND_CODE = 20012
	RES_ERR_DATABASE_QUERY_CODE  = 20015
	RES_ERR_DATABASE_ISSUE_CODE  = 20016
	RES_ERR_INTERNAL_SERVER_CODE = 20017

	// Finance domain (60000-60008)
	RES_ERR_ACCOUNT_NOT_FOUND_CODE      = 60000
	RES_ERR_ACCOUNT_FROZEN_CODE         = 60001
	RES_ERR_INSUFFICIENT_BALANCE_CODE   = 60002
	RES_ERR_INVALID_AMOUNT_CODE         = 60003
	RES_ERR_TRANSACTION_FAILED_CODE     = 60004
	RES_ERR_PAYMENT_FAILED_CODE         = 60005
	RES_ERR_PAYMENT_NOT_FOUND_CODE      = 60006
	RES_ERR_UNSUPPORTED_CURRENCY_CODE   = 60007
	RES_ERR_DEPOSIT_LIMIT_EXCEEDED_CODE = 60008
)

// Error keys
const (
	RES_ERR_INVALID_INPUT_KEY   = "res_err_invalid_input"
	RES_ERR_INVALID_PAYLOAD_KEY = "res_err_invalid_payload"
	RES_ERR_UNAUTHORIZED_KEY    = "res_err_unauthorized"
	RES_ERR_FORBIDDEN_KEY       = "res_err_forbidden"
	RES_ERR_ROUTE_NOT_FOUND_KEY = "res_err_route_not_found"
	RES_ERR_DATABASE_QUERY_KEY  = "res_err_database_query"
	RES_ERR_DATABASE_ISSUE_KEY  = "res_err_database_issue"
	RES_ERR_INTERNAL_SERVER_KEY = "res_err_internal_server"

	RES_ERR_ACCOUNT_NOT_FOUND_KEY      = "res_err_account_not_found"
	RES_ERR_ACCOUNT_FROZEN_KEY         = "res_err_account_frozen"
	RES_ERR_INSUFFICIENT_BALANCE_KEY   = "res_err_insufficient_balance"
	RES_ERR_INVALID_AMOUNT_KEY         = "res_err_invalid_amount"
	RES_ERR_TRANSACTION_FAILED_KEY     = "res_err_transaction_failed"
	RES_ERR_PAYMENT_FAILED_KEY         = "res_err_payment_failed"
	RES_ERR_PAYMENT_NOT_FOUND_KEY      = "res_err_payment_not_found"
	RES_ERR_UNSUPPORTED_CURRENCY_KEY   = "res_err_unsupported_currency"
	RES_ERR_DEPOSIT_LIMIT_EXCEEDED_KEY = "res_err_deposit_limit_exceeded"
)

// Error templates
var (
	RES_ERR_INVALID_INPUT = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusBadRequest,
		Code:       RES_ERR_INVALID_INPUT_CODE,
		Key:        RES_ERR_INVALID_INPUT_KEY,
		Message:    "Input invalid.",
	}

	RES_ERR_INVALID_PAYLOAD = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusBadRequest,
		Code:       RES_ERR_INVALID_PAYLOAD_CODE,
		Key:        RES_ERR_INVALID_PAYLOAD_KEY,
		Message:    "Payload invalid.",
	}

	RES_ERR_UNAUTHORIZED = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusUnauthorized,
		Code:       RES_ERR_UNAUTHORIZED_CODE,
		Key:        RES_ERR_UNAUTHORIZED_KEY,
		Message:    "Unauthorized.",
	}

	RES_ERR_FORBIDDEN = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusForbidden,
		Code:       RES_ERR_FORBIDDEN_CODE,
		Key:        RES_ERR_FORBIDDEN_KEY,
		Message:    "Forbidden.",
	}

	RES_ERR_ROUTE_NOT_FOUND = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusNotFound,
		Code:       RES_ERR_ROUTE_NOT_FOUND_CODE,
		Key:        RES_ERR_ROUTE_NOT_FOUND_KEY,
		Message:    "Requested endpoint not found.",
	}

	RES_ERR_DATABASE_QUERY = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusInternalServerError,
		Code:       RES_ERR_DATABASE_QUERY_CODE,
		Key:        RES_ERR_DATABASE_QUERY_KEY,
		Message:    "Error querying database, please try again.",
	}

	RES_ERR_DATABASE_ISSUE = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusInternalServerError,
		Code:       RES_ERR_DATABASE_ISSUE_CODE,
		Key:        RES_ERR_DATABASE_ISSUE_KEY,
		Message:    "Database issue, please try again.",
	}

	RES_ERR_INTERNAL_SERVER = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusInternalServerError,
		Code:       RES_ERR_INTERNAL_SERVER_CODE,
		Key:        RES_ERR_INTERNAL_SERVER_KEY,
		Message:    "Something went wrong.",
	}

	RES_ERR_ACCOUNT_NOT_FOUND = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusNotFound,
		Code:       RES_ERR_ACCOUNT_NOT_FOUND_CODE,
		Key:        RES_ERR_ACCOUNT_NOT_FOUND_KEY,
		Message:    "Wallet account not found.",
	}

	RES_ERR_ACCOUNT_FROZEN = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusForbidden,
		Code:       RES_ERR_ACCOUNT_FROZEN_CODE,
		Key:        RES_ERR_ACCOUNT_FROZEN_KEY,
		Message:    "Account is frozen.",
	}

	RES_ERR_INSUFFICIENT_BALANCE = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusBadRequest,
		Code:       RES_ERR_INSUFFICIENT_BALANCE_CODE,
		Key:        RES_ERR_INSUFFICIENT_BALANCE_KEY,
		Message:    "Insufficient balance.",
	}

	RES_ERR_INVALID_AMOUNT = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusBadRequest,
		Code:       RES_ERR_INVALID_AMOUNT_CODE,
		Key:        RES_ERR_INVALID_AMOUNT_KEY,
		Message:    "Amount must be a positive number.",
	}

	RES_ERR_TRANSACTION_FAILED = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusInternalServerError,
		Code:       RES_ERR_TRANSACTION_FAILED_CODE,
		Key:        RES_ERR_TRANSACTION_FAILED_KEY,
		Message:    "Transaction could not complete.",
	}

	RES_ERR_PAYMENT_FAILED = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusBadGateway,
		Code:       RES_ERR_PAYMENT_FAILED_CODE,
		Key:        RES_ERR_PAYMENT_FAILED_KEY,
		Message:    "Payment provider returned an error.",
	}

	RES_ERR_PAYMENT_NOT_FOUND = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusNotFound,
		Code:       RES_ERR_PAYMENT_NOT_FOUND_CODE,
		Key:        RES_ERR_PAYMENT_NOT_FOUND_KEY,
		Message:    "Payment record not found.",
	}

	RES_ERR_UNSUPPORTED_CURRENCY = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusBadRequest,
		Code:       RES_ERR_UNSUPPORTED_CURRENCY_CODE,
		Key:        RES_ERR_UNSUPPORTED_CURRENCY_KEY,
		Message:    "Currency code not recognized.",
	}

	RES_ERR_DEPOSIT_LIMIT_EXCEEDED = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusBadRequest,
		Code:       RES_ERR_DEPOSIT_LIMIT_EXCEEDED_CODE,
		Key:        RES_ERR_DEPOSIT_LIMIT_EXCEEDED_KEY,
		Message:    "Deposit amount exceeds limit.",
	}
)
