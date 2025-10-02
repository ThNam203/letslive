import { ApiResponse } from "@/types/fetch-response";
import GLOBAL from "../global";
import { FetchOptions } from "../types/fetch-options";

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
): Promise<T> => {
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

            // Retry the original request
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

async function validateAndParseResponse<T>(response: Response): Promise<T> {
    const contentType = response.headers.get("content-type");

    if (response.status === 204 || !contentType) {
        return {} as T;
    }

    if (contentType.includes("text/plain")) {
        return (await response.text()) as T;
    }

    const data = await response.json();
    return data as T;
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
