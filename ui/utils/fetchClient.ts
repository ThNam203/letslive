import { ApiResponse } from "@/types/fetch-response";
import GLOBAL from "../global";
import { FetchOptions } from "../types/fetch-options";

const DEFAULT_TIMEOUT_MS = 15000; // 15 seconds

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

    // Setup timeout controller
    let controller: AbortController | undefined;
    let timeoutId: NodeJS.Timeout | undefined;

    if (!options.disableTimeout) {
        const timeoutMs = options.timeoutMs ?? DEFAULT_TIMEOUT_MS; // default 15s
        controller = new AbortController();
        timeoutId = setTimeout(() => controller!.abort(), timeoutMs);
    }

    try {
        const response = await fetch(url, {
            credentials: "include",
            ...options,
            headers,
            signal: controller?.signal,
        });

        if (!response.ok) {
            if (response.status === 401 && !shouldSkipRefresh(url)) {
                await refreshToken();

                const retryResponse = await fetch(url, {
                    ...options,
                    credentials: "include",
                    headers,
                    signal: controller?.signal, // timeout
                });

                return validateAndParseResponse<T>(retryResponse);
            }
        }

        return validateAndParseResponse<T>(response);
    } catch (err: any) {
        if (err.name === "AbortError") {
            throw new Error(
                `request timed out after ${options.timeoutMs ?? DEFAULT_TIMEOUT_MS} ms`,
            );
        }
        throw err;
    } finally {
        if (timeoutId) clearTimeout(timeoutId);
    }
};

async function validateAndParseResponse<T>(
    response: Response,
): Promise<WithStatusCode<T>> {
    const contentType = response.headers.get("content-type");

    if (response.status === 204 || !contentType) {
        return { statusCode: response.status } as WithStatusCode<T>;
    }

    if (contentType.includes("text/plain")) {
        const text = await response.text();
        return { ...(text as unknown as T), statusCode: response.status };
    }

    const data = await response.json();
    return { ...data, statusCode: response.status };
}

const REFRESH_EXCLUDE_PATHS = [
    "/auth/login",
    "/auth/register",
    "/auth/refresh-token",
    "/auth/logout",
];

function shouldSkipRefresh(url: string): boolean {
    try {
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

function hasRefreshToken(): boolean {
    return (
        document.cookie.match(/^(.*;)?\s*REFRESH_TOKEN\s*=\s*[^;]+(.*)?$/) !=
        null
    );
}

function hasAccessToken(): boolean {
    return (
        document.cookie.match(/^(.*;)?\s*ACCESS_TOKEN\s*=\s*[^;]+(.*)?$/) !=
        null
    );
}
