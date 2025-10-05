import { ApiResponse } from "@/types/fetch-response";
import GLOBAL from "../global";
import { FetchOptions } from "../types/fetch-options";

type WithStatusCode<T> = T & { statusCode: number };

// Singleton promise for token refresh
let refreshTokenPromise: Promise<ApiResponse<void>> | null = null;

const refreshToken = async (): Promise<ApiResponse<void>> => {
    if (!refreshTokenPromise) {
        refreshTokenPromise = (async () => {
            const refreshResponse = await fetch(
                `${GLOBAL.API_URL}/auth/refresh-token`,
                { method: "POST", credentials: "include" },
            );

            return await refreshResponse
                .json()
                .catch(() => null)
                .finally(() => {
                    refreshTokenPromise = null;
                });
        })();
    }
    return refreshTokenPromise;
};

// handle exception outside
export const fetchClient = async <T>(
    url: string,
    options: FetchOptions = {},
): Promise<WithStatusCode<T>> => {
    if (!url.startsWith("http")) {
        url = GLOBAL.API_URL + url;
    }

    const defaultHeaders: Record<string, string> = {
        "Cache-Control": "no-store",
    };

    const headers = {
        ...defaultHeaders,
        ...(options.headers ?? {}),
    };

    const response = await fetch(url, {
        credentials: "include",
        ...options,
        headers,
    });

    if (!response.ok) {
        if (response.status === 401 && !shouldSkipRefresh(url)) {
            await refreshToken();

            const retryResponse = await fetch(url, {
                ...options,
                credentials: "include",
                headers,
                body: options.body,
            });

            return validateAndParseResponse<T>(retryResponse);
        }
    }

    return validateAndParseResponse<T>(response);
};

async function validateAndParseResponse<T>(response: Response): Promise<WithStatusCode<T>> {
    const contentType = response.headers.get("content-type");

    if (response.status === 204 || !contentType) {
        return { statusCode: response.status } as WithStatusCode<T>;
    }

    // TODO: should not hit this if
    if (contentType.includes("text/plain")) {
        const text = await response.text();
        return { ...(text as unknown as T), statusCode: response.status };
    }

    const data = await response.json();

    // TODO: temp fix for messages endpoint
    if (Array.isArray(data)) {
        return { data: data, statusCode: response.status } as any;
    }
    return { ...data, statusCode: response.status };
}

// URLs that should NOT trigger refresh
const REFRESH_EXCLUDE_PATHS = [
    "/auth/login",
    "/auth/register",
    "/auth/refresh-token",
    "/auth/logout",
];

function shouldSkipRefresh(url: string): boolean {
    try {
        // Normalize relative vs absolute
        const u = url.startsWith("http")
            ? new URL(url)
            : new URL(url, GLOBAL.API_URL);
        return REFRESH_EXCLUDE_PATHS.some((path) =>
            u.pathname.startsWith(path),
        );
    } catch {
        return false; // if URL parsing fails, allow refresh just in case
    }
}

// TODO: better api-call handler
function hasRefreshToken(): boolean {
    return document.cookie.match(/^(.*;)?\s*REFRESH_TOKEN\s*=\s*[^;]+(.*)?$/) != null;
}

function hasAccessToken(): boolean {
    return document.cookie.match(/^(.*;)?\s*ACCESS_TOKEN\s*=\s*[^;]+(.*)?$/) != null;
}