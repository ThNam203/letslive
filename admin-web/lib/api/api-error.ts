import { ApiResponse } from "@/types/fetch-response";

export class ApiError extends Error {
    key: string;
    requestId: string;
    statusCode: number;

    constructor(
        res: Pick<ApiResponse<unknown>, "key" | "requestId" | "message" | "statusCode">,
    ) {
        super(res.message);
        this.name = "ApiError";
        this.key = res.key;
        this.requestId = res.requestId;
        this.statusCode = res.statusCode;
    }
}

export function unwrapResponse<T>(res: ApiResponse<T>): T {
    if (!res.success) {
        throw new ApiError(res);
    }
    return res.data as T;
}
