package response

import "net/http"

// Error codes
const (
	RES_ERR_INVALID_INPUT_CODE   = 20000
	RES_ERR_INVALID_PAYLOAD_CODE = 20001
	RES_ERR_UNAUTHORIZED_CODE    = 20005
	RES_ERR_FORBIDDEN_CODE       = 20008
	RES_ERR_ROUTE_NOT_FOUND_CODE = 20012
	RES_ERR_IMAGE_TOO_LARGE_CODE = 30002
	RES_ERR_DATABASE_QUERY_CODE  = 20015
	RES_ERR_DATABASE_ISSUE_CODE  = 20016
	RES_ERR_INTERNAL_SERVER_CODE = 20017

	RES_ERR_LIVESTREAM_UPDATE_AFTER_ENDED_CODE = 40000
	RES_ERR_LIVESTREAM_NOT_FOUND_CODE          = 40001
	RES_ERR_VOD_NOT_FOUND_CODE                 = 40002
	RES_ERR_END_ALREADY_ENDED_LIVESTREAM_CODE  = 40003
	RES_ERR_QUERY_SCAN_FAILED_CODE             = 40004
	RES_ERR_LIVESTREAM_CREATE_FAILED_CODE      = 40005
	RES_ERR_LIVESTREAM_UPDATE_FAILED_CODE      = 40006
	RES_ERR_VOD_CREATE_FAILED_CODE             = 40007
	RES_ERR_VOD_UPDATE_FAILED_CODE             = 40008
)

// Error keys
const (
	RES_ERR_INVALID_INPUT_KEY   = "res_err_invalid_input"
	RES_ERR_INVALID_PAYLOAD_KEY = "res_err_invalid_payload"
	RES_ERR_UNAUTHORIZED_KEY    = "res_err_unauthorized"
	RES_ERR_FORBIDDEN_KEY       = "res_err_forbidden"
	RES_ERR_ROUTE_NOT_FOUND_KEY = "res_err_route_not_found"
	RES_ERR_IMAGE_TOO_LARGE_KEY = "res_err_image_too_large"
	RES_ERR_DATABASE_QUERY_KEY  = "res_err_database_query"
	RES_ERR_DATABASE_ISSUE_KEY  = "res_err_database_issue"
	RES_ERR_INTERNAL_SERVER_KEY = "res_err_internal_server"

	RES_ERR_LIVESTREAM_UPDATE_AFTER_ENDED_KEY = "res_err_livestream_update_after_ended"
	RES_ERR_LIVESTREAM_NOT_FOUND_KEY          = "res_err_livestream_not_found"
	RES_ERR_VOD_NOT_FOUND_KEY                 = "res_err_vod_not_found"
	RES_ERR_END_ALREADY_ENDED_LIVESTREAM_KEY  = "res_err_end_already_ended_livestream"
	RES_ERR_LIVESTREAM_CREATE_FAILED_KEY      = "res_err_livestream_create_failed"
	RES_ERR_LIVESTREAM_UPDATE_FAILED_KEY      = "res_err_livestream_update_failed"
	RES_ERR_VOD_CREATE_FAILED_KEY             = "res_err_vod_create_failed"
	RES_ERR_VOD_UPDATE_FAILED_KEY             = "res_err_vod_update_failed"
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

	RES_ERR_LIVESTREAM_UPDATE_AFTER_ENDED = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusBadRequest,
		Code:       RES_ERR_LIVESTREAM_UPDATE_AFTER_ENDED_CODE,
		Key:        RES_ERR_LIVESTREAM_UPDATE_AFTER_ENDED_KEY,
		Message:    "Failed to update, the livestream has ended.",
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

	RES_ERR_LIVESTREAM_NOT_FOUND = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusNotFound,
		Code:       RES_ERR_LIVESTREAM_NOT_FOUND_CODE,
		Key:        RES_ERR_LIVESTREAM_NOT_FOUND_KEY,
		Message:    "Livestream not found.",
	}

	RES_ERR_VOD_NOT_FOUND = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusNotFound,
		Code:       RES_ERR_VOD_NOT_FOUND_CODE,
		Key:        RES_ERR_VOD_NOT_FOUND_KEY,
		Message:    "VOD not found.",
	}

	RES_ERR_ROUTE_NOT_FOUND = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusNotFound,
		Code:       RES_ERR_ROUTE_NOT_FOUND_CODE,
		Key:        RES_ERR_ROUTE_NOT_FOUND_KEY,
		Message:    "Requested endpoint not found.",
	}

	RES_ERR_END_ALREADY_ENDED_LIVESTREAM = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusConflict,
		Code:       RES_ERR_END_ALREADY_ENDED_LIVESTREAM_CODE,
		Key:        RES_ERR_END_ALREADY_ENDED_LIVESTREAM_KEY,
		Message:    "The livestream has already been ended.",
	}

	RES_ERR_IMAGE_TOO_LARGE = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusRequestEntityTooLarge,
		Code:       RES_ERR_IMAGE_TOO_LARGE_CODE,
		Key:        RES_ERR_IMAGE_TOO_LARGE_KEY,
		Message:    "Image exceeds 10mb limit.",
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
		Message:    "Database issue, please try again.",
	}

	RES_ERR_INTERNAL_SERVER = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusInternalServerError,
		Code:       RES_ERR_INTERNAL_SERVER_CODE,
		Key:        RES_ERR_INTERNAL_SERVER_KEY,
		Message:    "Something went wrong.",
	}

	RES_ERR_LIVESTREAM_CREATE_FAILED = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusInternalServerError,
		Code:       RES_ERR_LIVESTREAM_CREATE_FAILED_CODE,
		Key:        RES_ERR_LIVESTREAM_CREATE_FAILED_KEY,
		Message:    "Failed to create livestream record.",
	}

	RES_ERR_LIVESTREAM_UPDATE_FAILED = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusInternalServerError,
		Code:       RES_ERR_LIVESTREAM_UPDATE_FAILED_CODE,
		Key:        RES_ERR_LIVESTREAM_UPDATE_FAILED_KEY,
		Message:    "Failed to update livestream record.",
	}

	RES_ERR_VOD_CREATE_FAILED = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusInternalServerError,
		Code:       RES_ERR_VOD_CREATE_FAILED_CODE,
		Key:        RES_ERR_VOD_CREATE_FAILED_KEY,
		Message:    "Failed to create vod record.",
	}

	RES_ERR_VOD_UPDATE_FAILED = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusInternalServerError,
		Code:       RES_ERR_VOD_UPDATE_FAILED_CODE,
		Key:        RES_ERR_VOD_UPDATE_FAILED_KEY,
		Message:    "Failed to update vod record.",
	}
)
