export type Meta = {
    page?: number;
    page_size?: number;
    total?: number;
};

export type ErrorDetail = Record<string, any>;

export type ErrorDetails = ErrorDetail[];

export type ApiResponse<T> = {
    requestId: string; // request/trace id (from header)
    success: boolean;
    statusCode: number; // got from the fetch's response
    code: number; // business-level code
    key: string; // i18n key
    message: string; // english default
    data?: T;
    meta?: Meta; // pagination/filter info
    errorDetails?: ErrorDetails; // validation details, etc.
};

// --- API codes
export enum ApiCode {
    // General / Auth (200xx)
    RES_ERR_INVALID_INPUT = 20000,
    RES_ERR_INVALID_PAYLOAD = 20001,
    RES_ERR_AUTH_ALREADY_EXISTS = 20002,
    RES_ERR_CAPTCHA_FAILED = 20003,
    RES_ERR_PASSWORD_NOT_MATCH = 20004,
    RES_ERR_UNAUTHORIZED = 20005,
    RES_ERR_SIGN_UP_OTP_EXPIRED = 20006,
    RES_ERR_EMAIL_OR_PASSWORD_INCORRECT = 20007,
    RES_ERR_FORBIDDEN = 20008,
    RES_ERR_AUTH_NOT_FOUND = 20009,
    RES_ERR_REFRESH_TOKEN_NOT_FOUND = 20010,
    RES_ERR_SIGN_UP_OTP_NOT_FOUND = 20011,
    RES_ERR_ROUTE_NOT_FOUND = 20012,
    RES_ERR_SIGN_UP_OTP_ALREADY_USED = 20013,
    RES_ERR_FAILED_TO_CREATE_SIGN_UP_OTP = 20014,
    RES_ERR_DATABASE_QUERY = 20015,
    RES_ERR_DATABASE_ISSUE = 20016,
    RES_ERR_INTERNAL_SERVER = 20017,
    RES_ERR_FAILED_TO_SEND_VERIFICATION = 20018,

    // User (300xx)
    RES_ERR_USER_NOT_FOUND = 30000,
    RES_ERR_IMAGE_TOO_LARGE = 30001,

    // Livestream / VOD (400xx)
    RES_ERR_LIVESTREAM_UPDATE_AFTER_ENDED = 40000,
    RES_ERR_LIVESTREAM_NOT_FOUND = 40001,
    RES_ERR_VOD_NOT_FOUND = 40002,
    RES_ERR_END_ALREADY_ENDED_LIVESTREAM = 40003,
    RES_ERR_QUERY_SCAN_FAILED = 40004,
    RES_ERR_LIVESTREAM_CREATE_FAILED = 40005,
    RES_ERR_LIVESTREAM_UPDATE_FAILED = 40006,
    RES_ERR_VOD_CREATE_FAILED = 40007,
    RES_ERR_VOD_UPDATE_FAILED = 40008,
}

// --- i18n keys
export enum ApiKey {
    RES_ERR_INVALID_INPUT = "res_err_invalid_input",
    RES_ERR_INVALID_PAYLOAD = "res_err_invalid_payload",
    RES_ERR_AUTH_ALREADY_EXISTS = "res_err_auth_already_exists",
    RES_ERR_CAPTCHA_FAILED = "res_err_captcha_failed",
    RES_ERR_PASSWORD_NOT_MATCH = "res_err_password_not_match",
    RES_ERR_UNAUTHORIZED = "res_err_unauthorized",
    RES_ERR_SIGN_UP_OTP_EXPIRED = "res_err_sign_up_otp_expired",
    RES_ERR_EMAIL_OR_PASSWORD_INCORRECT = "res_err_email_or_password_incorrect",
    RES_ERR_FORBIDDEN = "res_err_forbidden",
    RES_ERR_AUTH_NOT_FOUND = "res_err_auth_not_found",
    RES_ERR_REFRESH_TOKEN_NOT_FOUND = "res_err_refresh_token_not_found",
    RES_ERR_SIGN_UP_OTP_NOT_FOUND = "res_err_sign_up_otp_not_found",
    RES_ERR_ROUTE_NOT_FOUND = "res_err_route_not_found",
    RES_ERR_SIGN_UP_OTP_ALREADY_USED = "res_err_sign_up_otp_already_used",
    RES_ERR_FAILED_TO_CREATE_SIGN_UP_OTP = "res_err_failed_to_create_sign_up_otp",
    RES_ERR_DATABASE_QUERY = "res_err_database_query",
    RES_ERR_DATABASE_ISSUE = "res_err_database_issue",
    RES_ERR_INTERNAL_SERVER = "res_err_internal_server",
    RES_ERR_FAILED_TO_SEND_VERIFICATION = "res_err_failed_to_send_verification",

    RES_ERR_USER_NOT_FOUND = "res_err_user_not_found",
    RES_ERR_IMAGE_TOO_LARGE = "res_err_image_too_large",

    RES_ERR_LIVESTREAM_UPDATE_AFTER_ENDED = "res_err_livestream_update_after_ended",
    RES_ERR_LIVESTREAM_NOT_FOUND = "res_err_livestream_not_found",
    RES_ERR_VOD_NOT_FOUND = "res_err_vod_not_found",
    RES_ERR_END_ALREADY_ENDED_LIVESTREAM = "res_err_end_already_ended_livestream",
    RES_ERR_QUERY_SCAN_FAILED = "res_err_query_scan_failed",
    RES_ERR_LIVESTREAM_CREATE_FAILED = "res_err_livestream_create_failed",
    RES_ERR_LIVESTREAM_UPDATE_FAILED = "res_err_livestream_update_failed",
    RES_ERR_VOD_CREATE_FAILED = "res_err_vod_create_failed",
    RES_ERR_VOD_UPDATE_FAILED = "res_err_vod_update_failed",
}
