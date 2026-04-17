import { ApiResponse, Meta } from "@/types/fetch-response";
import GLOBAL from "@/global";
import { HttpResponse } from "msw";

let _reqSeq = 1;
const requestId = () => `mock-req-${_reqSeq++}`;

// All helpers return the native `Response` type so MSW handlers can freely
// mix return types without TypeScript complaining about body-type mismatches.

export function ok<T>(data: T, meta?: Meta): Response {
    const body: ApiResponse<T> = {
        requestId: requestId(),
        success: true,
        statusCode: 200,
        code: 0,
        key: "res_success",
        message: "OK",
        data,
        meta,
    };
    return HttpResponse.json(body, { status: 200 });
}

export function created<T>(data: T): Response {
    const body: ApiResponse<T> = {
        requestId: requestId(),
        success: true,
        statusCode: 201,
        code: 0,
        key: "res_success",
        message: "Created",
        data,
    };
    return HttpResponse.json(body, { status: 201 });
}

export function noContent(): Response {
    return new HttpResponse(null, { status: 204 });
}

export function notFound(key: string, message: string): Response {
    const body: ApiResponse<null> = {
        requestId: requestId(),
        success: false,
        statusCode: 404,
        code: 0,
        key,
        message,
    };
    return HttpResponse.json(body, { status: 404 });
}

export function badRequest(key: string, message: string): Response {
    const body: ApiResponse<null> = {
        requestId: requestId(),
        success: false,
        statusCode: 400,
        code: 0,
        key,
        message,
    };
    return HttpResponse.json(body, { status: 400 });
}

export function unauthorized(): Response {
    const body: ApiResponse<null> = {
        requestId: requestId(),
        success: false,
        statusCode: 401,
        code: 20005,
        key: "res_err_unauthorized",
        message: "Unauthorized",
    };
    return HttpResponse.json(body, { status: 401 });
}

/** Same origin the app uses in fetchClient (avoids env-inlining mismatches vs handlers). */
export const API_BASE = GLOBAL.API_URL;
