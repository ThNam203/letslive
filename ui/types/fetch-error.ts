export type ErrorResponse = {
    id: string;
    message: string;
    code: string; // general error code
    key: string; // i18n key
};

export class FetchError extends Error {
    id: string;
    status?: number;
    response?: any;
    payload?: any;

    constructor(
        id: string,
        code: string,
        message: string,
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
