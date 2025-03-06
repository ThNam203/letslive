export type APIError = {
    message: string
    statusCode: number
}

export const ApiErrors = {
    INVALID_INPUT: {
        message: 'Invalid input ',
        statusCode: 400
    } as APIError,
    INVALID_PATH: {
        message: 'Invalid path',
        statusCode: 401
    } as APIError,
    ROUTE_NOT_FOUND: {
        message: 'Route not found',
        statusCode: 404
    } as APIError,
    INTERNAL_SERVER_ERROR: {
        message: 'Internal server error',
        statusCode: 500
    } as APIError
} as const

export const createApiError = (message: string, statusCode: number): APIError => ({
    message,
    statusCode
})
