export type ErrorDetail = Record<string, unknown>;
export type ErrorDetails = ErrorDetail[];

export type ApiResponse<T> = {
    requestId: string;
    success: boolean;
    statusCode: number;
    code: number;
    key: string;
    message: string;
    data?: T;
    errorDetails?: ErrorDetails;
};
