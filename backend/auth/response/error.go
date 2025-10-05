package serviceresponse

import "net/http"

const (
	RES_ERR_INVALID_INPUT_CODE                = 20000
	RES_ERR_INVALID_PAYLOAD_CODE              = 20001
	RES_ERR_AUTH_ALREADY_EXISTS_CODE          = 20002
	RES_ERR_CAPTCHA_FAILED_CODE               = 20003
	RES_ERR_PASSWORD_NOT_MATCH_CODE           = 20004
	RES_ERR_UNAUTHORIZED_CODE                 = 20005
	RES_ERR_SIGN_UP_OTP_EXPIRED_CODE          = 20006
	RES_ERR_EMAIL_OR_PASSWORD_INCORRECT_CODE  = 20007
	RES_ERR_FORBIDDEN_CODE                    = 20008
	RES_ERR_AUTH_NOT_FOUND_CODE               = 20009
	RES_ERR_REFRESH_TOKEN_NOT_FOUND_CODE      = 20010
	RES_ERR_SIGN_UP_OTP_NOT_FOUND_CODE        = 20011
	RES_ERR_ROUTE_NOT_FOUND_CODE              = 20012
	RES_ERR_SIGN_UP_OTP_ALREADY_USED_CODE     = 20013
	RES_ERR_FAILED_TO_CREATE_SIGN_UP_OTP_CODE = 20014
	RES_ERR_DATABASE_QUERY_CODE               = 20015
	RES_ERR_DATABASE_ISSUE_CODE               = 20016
	RES_ERR_INTERNAL_SERVER_CODE              = 20017
	RES_ERR_FAILED_TO_SEND_VERIFICATION_CODE  = 20018
)

const (
	RES_ERR_INVALID_INPUT_KEY                = "res_err_invalid_input"
	RES_ERR_INVALID_PAYLOAD_KEY              = "res_err_invalid_payload"
	RES_ERR_AUTH_ALREADY_EXISTS_KEY          = "res_err_auth_already_exists"
	RES_ERR_CAPTCHA_FAILED_KEY               = "res_err_captcha_failed"
	RES_ERR_PASSWORD_NOT_MATCH_KEY           = "res_err_password_not_match"
	RES_ERR_UNAUTHORIZED_KEY                 = "res_err_unauthorized"
	RES_ERR_SIGN_UP_OTP_EXPIRED_KEY          = "res_err_sign_up_otp_expired"
	RES_ERR_EMAIL_OR_PASSWORD_INCORRECT_KEY  = "res_err_email_or_password_incorrect"
	RES_ERR_FORBIDDEN_KEY                    = "res_err_forbidden"
	RES_ERR_AUTH_NOT_FOUND_KEY               = "res_err_auth_not_found"
	RES_ERR_REFRESH_TOKEN_NOT_FOUND_KEY      = "res_err_refresh_token_not_found"
	RES_ERR_SIGN_UP_OTP_NOT_FOUND_KEY        = "res_err_sign_up_otp_not_found"
	RES_ERR_ROUTE_NOT_FOUND_KEY              = "res_err_route_not_found"
	RES_ERR_SIGN_UP_OTP_ALREADY_USED_KEY     = "res_err_sign_up_otp_already_used"
	RES_ERR_FAILED_TO_CREATE_SIGN_UP_OTP_KEY = "res_err_failed_to_create_sign_up_otp"
	RES_ERR_DATABASE_QUERY_KEY               = "res_err_database_query"
	RES_ERR_DATABASE_ISSUE_KEY               = "res_err_database_issue"
	RES_ERR_INTERNAL_SERVER_KEY              = "res_err_internal_server"
	RES_ERR_FAILED_TO_SEND_VERIFICATION_KEY  = "res_err_failed_to_send_verification"
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

	RES_ERR_AUTH_ALREADY_EXISTS = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusBadRequest,
		Code:       RES_ERR_AUTH_ALREADY_EXISTS_CODE,
		Key:        RES_ERR_AUTH_ALREADY_EXISTS_KEY,
		Message:    "Email has already been registered.",
	}

	RES_ERR_CAPTCHA_FAILED = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusBadRequest,
		Code:       RES_ERR_CAPTCHA_FAILED_CODE,
		Key:        RES_ERR_CAPTCHA_FAILED_KEY,
		Message:    "Failed to verify CAPTCHA, please try again.",
	}

	RES_ERR_PASSWORD_NOT_MATCH = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusBadRequest,
		Code:       RES_ERR_PASSWORD_NOT_MATCH_CODE,
		Key:        RES_ERR_PASSWORD_NOT_MATCH_KEY,
		Message:    "Old password does not match.",
	}

	RES_ERR_UNAUTHORIZED = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusUnauthorized,
		Code:       RES_ERR_UNAUTHORIZED_CODE,
		Key:        RES_ERR_UNAUTHORIZED_KEY,
		Message:    "Unauthorized.",
	}

	RES_ERR_SIGN_UP_OTP_EXPIRED = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusUnauthorized,
		Code:       RES_ERR_SIGN_UP_OTP_EXPIRED_CODE,
		Key:        RES_ERR_SIGN_UP_OTP_EXPIRED_KEY,
		Message:    "OTP code has expired, please issue a new one.",
	}

	RES_ERR_EMAIL_OR_PASSWORD_INCORRECT = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusUnauthorized,
		Code:       RES_ERR_EMAIL_OR_PASSWORD_INCORRECT_CODE,
		Key:        RES_ERR_EMAIL_OR_PASSWORD_INCORRECT_KEY,
		Message:    "Username or password incorrect.",
	}

	RES_ERR_FORBIDDEN = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusForbidden,
		Code:       RES_ERR_FORBIDDEN_CODE,
		Key:        RES_ERR_FORBIDDEN_KEY,
		Message:    "Forbidden.",
	}

	RES_ERR_AUTH_NOT_FOUND = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusNotFound,
		Code:       RES_ERR_AUTH_NOT_FOUND_CODE,
		Key:        RES_ERR_AUTH_NOT_FOUND_KEY,
		Message:    "Authentication credentials not found.",
	}

	RES_ERR_REFRESH_TOKEN_NOT_FOUND = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusNotFound,
		Code:       RES_ERR_REFRESH_TOKEN_NOT_FOUND_CODE,
		Key:        RES_ERR_REFRESH_TOKEN_NOT_FOUND_KEY,
		Message:    "Refresh token not found.",
	}

	RES_ERR_SIGN_UP_OTP_NOT_FOUND = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusNotFound,
		Code:       RES_ERR_SIGN_UP_OTP_NOT_FOUND_CODE,
		Key:        RES_ERR_SIGN_UP_OTP_NOT_FOUND_KEY,
		Message:    "OTP code not found.",
	}

	RES_ERR_ROUTE_NOT_FOUND = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusNotFound,
		Code:       RES_ERR_ROUTE_NOT_FOUND_CODE,
		Key:        RES_ERR_ROUTE_NOT_FOUND_KEY,
		Message:    "Requested endpoint not found.",
	}

	RES_ERR_SIGN_UP_OTP_ALREADY_USED = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusConflict,
		Code:       RES_ERR_SIGN_UP_OTP_ALREADY_USED_CODE,
		Key:        RES_ERR_SIGN_UP_OTP_ALREADY_USED_KEY,
		Message:    "The OTP has already been used.",
	}

	RES_ERR_FAILED_TO_CREATE_SIGN_UP_OTP = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusConflict,
		Code:       RES_ERR_FAILED_TO_CREATE_SIGN_UP_OTP_CODE,
		Key:        RES_ERR_FAILED_TO_CREATE_SIGN_UP_OTP_KEY,
		Message:    "Failed to generate the OTP, please try again later.",
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

	RES_ERR_FAILED_TO_SEND_VERIFICATION = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusInternalServerError,
		Code:       RES_ERR_FAILED_TO_SEND_VERIFICATION_CODE,
		Key:        RES_ERR_FAILED_TO_SEND_VERIFICATION_KEY,
		Message:    "Failed to send email verification, please try again later.",
	}
)
