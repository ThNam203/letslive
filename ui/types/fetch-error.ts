export type ErrorResponse = {
    id: string;
    message: string;
    statusCode: number;
};

export class FetchError extends Error {
    id: string;
    status?: number;
    response?: any;
    isClientError?: boolean;
    isServerError?: boolean;
    isNetworkError?: boolean;
    payload?: any;

    constructor(
        id: string,
        message: string,
        options?: {
            status?: number;
            response?: any;
            isClientError?: boolean;
            isServerError?: boolean;
            isNetworkError?: boolean;
            payload?: any;
        }
    ) {
        super(message);
        this.id = id;
        this.name = "FetchError";
        this.status = options?.status;
        this.response = options?.response;
        this.isClientError = options?.isClientError;
        this.isServerError = options?.isServerError;
        this.isNetworkError = options?.isNetworkError;
        this.payload = options?.payload;

        // Ensure the prototype chain is properly set
        Object.setPrototypeOf(this, FetchError.prototype);
    }
}
