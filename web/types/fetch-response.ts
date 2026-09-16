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
