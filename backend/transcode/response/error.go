package response

import "net/http"

const (
	RES_ERR_INVALID_INPUT_CODE   = 20000
	RES_ERR_INVALID_PAYLOAD_CODE = 20001
	RES_ERR_UNAUTHORIZED_CODE    = 20005
	RES_ERR_FORBIDDEN_CODE       = 20008
	RES_ERR_ROUTE_NOT_FOUND_CODE = 20012
	RES_ERR_DATABASE_QUERY_CODE  = 20015
	RES_ERR_DATABASE_ISSUE_CODE  = 20016
	RES_ERR_INTERNAL_SERVER_CODE = 20017
)

const (
	RES_ERR_INVALID_INPUT_KEY   = "res_err_invalid_input"
	RES_ERR_INVALID_PAYLOAD_KEY = "res_err_invalid_payload"
	RES_ERR_UNAUTHORIZED_KEY    = "res_err_unauthorized"
	RES_ERR_FORBIDDEN_KEY       = "res_err_forbidden"
	RES_ERR_ROUTE_NOT_FOUND_KEY = "res_err_route_not_found"
	RES_ERR_DATABASE_QUERY_KEY  = "res_err_database_query"
	RES_ERR_DATABASE_ISSUE_KEY  = "res_err_database_issue"
	RES_ERR_INTERNAL_SERVER_KEY = "res_err_internal_server"
)

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
		Message:    "Something went wrong.",
	}

	RES_ERR_DATABASE_ISSUE = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusInternalServerError,
		Code:       RES_ERR_DATABASE_ISSUE_CODE,
		Key:        RES_ERR_DATABASE_ISSUE_KEY,
		Message:    "Something went wrong.",
	}

	RES_ERR_INTERNAL_SERVER = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusInternalServerError,
		Code:       RES_ERR_INTERNAL_SERVER_CODE,
		Key:        RES_ERR_INTERNAL_SERVER_KEY,
		Message:    "Something went wrong.",
	}
)
