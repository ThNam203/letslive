// --- Error Codes ---
export const RES_ERR_INVALID_INPUT_CODE = 20000
export const RES_ERR_INVALID_PAYLOAD_CODE = 20001
export const RES_ERR_UNAUTHORIZED_CODE = 20005
export const RES_ERR_FORBIDDEN_CODE = 20008
export const RES_ERR_ROUTE_NOT_FOUND_CODE = 20012
export const RES_ERR_IMAGE_TOO_LARGE_CODE = 30002
export const RES_ERR_DATABASE_QUERY_CODE = 20015
export const RES_ERR_DATABASE_ISSUE_CODE = 20016
export const RES_ERR_INTERNAL_SERVER_CODE = 20017

export const RES_ERR_ROOM_NOT_FOUND_CODE = 50018

// --- Error Keys ---
export const RES_ERR_INVALID_INPUT_KEY = 'res_err_invalid_input'
export const RES_ERR_INVALID_PAYLOAD_KEY = 'res_err_invalid_payload'
export const RES_ERR_UNAUTHORIZED_KEY = 'res_err_unauthorized'
export const RES_ERR_FORBIDDEN_KEY = 'res_err_forbidden'
export const RES_ERR_ROUTE_NOT_FOUND_KEY = 'res_err_route_not_found'
export const RES_ERR_IMAGE_TOO_LARGE_KEY = 'res_err_image_too_large'
export const RES_ERR_DATABASE_QUERY_KEY = 'res_err_database_query'
export const RES_ERR_DATABASE_ISSUE_KEY = 'res_err_database_issue'
export const RES_ERR_INTERNAL_SERVER_KEY = 'res_err_internal_server'
export const RES_ERR_ROOM_NOT_FOUND_KEY = 'res_err_room_not_found'

// --- Success Codes & Keys ---
export const RES_SUCC_OK_CODE = 100000
export const RES_SUCC_OK_KEY = 'res_succ_ok'

// --- HTTP Status Constants ---
export const HTTP_STATUS_OK = 200
export const HTTP_STATUS_BAD_REQUEST = 400
export const HTTP_STATUS_UNAUTHORIZED = 401
export const HTTP_STATUS_FORBIDDEN = 403
export const HTTP_STATUS_NOT_FOUND = 404
export const HTTP_STATUS_INTERNAL_SERVER_ERROR = 500

export interface ResponseTemplate {
    success: boolean
    statusCode: number
    code: number
    key: string
    message?: string
}

export const RESPONSE_TEMPLATES = {
    // --- Success Template ---
    RES_SUCC_OK: {
        success: true,
        statusCode: HTTP_STATUS_OK,
        code: RES_SUCC_OK_CODE,
        key: RES_SUCC_OK_KEY
    },

    // --- Error Templates ---
    RES_ERR_INVALID_INPUT: {
        success: false,
        statusCode: HTTP_STATUS_BAD_REQUEST,
        code: RES_ERR_INVALID_INPUT_CODE,
        key: RES_ERR_INVALID_INPUT_KEY
    },
    RES_ERR_INVALID_PAYLOAD: {
        success: false,
        statusCode: HTTP_STATUS_BAD_REQUEST,
        code: RES_ERR_INVALID_PAYLOAD_CODE,
        key: RES_ERR_INVALID_PAYLOAD_KEY
    },
    RES_ERR_UNAUTHORIZED: {
        success: false,
        statusCode: HTTP_STATUS_UNAUTHORIZED,
        code: RES_ERR_UNAUTHORIZED_CODE,
        key: RES_ERR_UNAUTHORIZED_KEY
    },
    RES_ERR_FORBIDDEN: {
        success: false,
        statusCode: HTTP_STATUS_FORBIDDEN,
        code: RES_ERR_FORBIDDEN_CODE,
        key: RES_ERR_FORBIDDEN_KEY
    },
    RES_ERR_ROUTE_NOT_FOUND: {
        success: false,
        statusCode: HTTP_STATUS_NOT_FOUND,
        code: RES_ERR_ROUTE_NOT_FOUND_CODE,
        key: RES_ERR_ROUTE_NOT_FOUND_KEY
    },
    RES_ERR_IMAGE_TOO_LARGE: {
        success: false,
        statusCode: HTTP_STATUS_BAD_REQUEST,
        code: RES_ERR_IMAGE_TOO_LARGE_CODE,
        key: RES_ERR_IMAGE_TOO_LARGE_KEY
    },
    RES_ERR_DATABASE_QUERY: {
        success: false,
        statusCode: HTTP_STATUS_INTERNAL_SERVER_ERROR,
        code: RES_ERR_DATABASE_QUERY_CODE,
        key: RES_ERR_DATABASE_QUERY_KEY
    },
    RES_ERR_DATABASE_ISSUE: {
        success: false,
        statusCode: HTTP_STATUS_INTERNAL_SERVER_ERROR,
        code: RES_ERR_DATABASE_ISSUE_CODE,
        key: RES_ERR_DATABASE_ISSUE_KEY
    },
    RES_ERR_INTERNAL_SERVER: {
        success: false,
        statusCode: HTTP_STATUS_INTERNAL_SERVER_ERROR,
        code: RES_ERR_INTERNAL_SERVER_CODE,
        key: RES_ERR_INTERNAL_SERVER_KEY
    },
    RES_ERR_ROOM_NOT_FOUND: {
        success: false,
        statusCode: HTTP_STATUS_NOT_FOUND,
        code: RES_ERR_ROOM_NOT_FOUND_CODE,
        key: RES_ERR_ROOM_NOT_FOUND_KEY,
        message: 'Room not found'
    }
} as const

// Optional: type of all templates
export type ResponseTemplateKey = keyof typeof RESPONSE_TEMPLATES

export interface Meta {
    page?: number
    page_size?: number
    total?: number
}

export type ErrorDetail = Record<string, any>

export type ErrorDetails = ErrorDetail[]

export interface Response<T> {
    requestId: string
    success: boolean
    statusCode: number
    code: number // business-level code
    key: string // i18n key
    message?: string // default English message
    data?: T | null // generic data
    meta?: Meta | null // pagination, filters, etc.
    errorDetails?: ErrorDetails | null
}

/**
 * Builds a Response<T> object based on a template.
 */
export function newResponseFromTemplate<T>(
    tpl: ResponseTemplate,
    data?: T | null,
    meta?: Meta | null,
    errorDetails?: ErrorDetails | null
): Response<T> {
    return {
        requestId: '', // can be filled by middleware later
        success: tpl.success,
        statusCode: tpl.statusCode,
        code: tpl.code,
        key: tpl.key,
        message: tpl.message,
        data,
        meta,
        errorDetails
    }
}
