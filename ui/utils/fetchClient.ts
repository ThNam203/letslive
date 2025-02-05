import { ErrorResponse, FetchError } from "@/types/fetch-error";
import { FetchOptions } from "@/types/fetch-options";

const BASE_URL = "http://localhost:8000";

// Singleton promise for token refresh
let refreshTokenPromise: Promise<void> | null = null;

const refreshToken = async (): Promise<void> => {
    if (!refreshTokenPromise) {
        refreshTokenPromise = (async () => {
            try {
                const refreshResponse = await fetch(
                    `${BASE_URL}/auth/refresh-token`,
                    { method: "POST", credentials: "include" }
                );

                
                const refreshRes = await refreshResponse.json().catch(() => null);
                if (!refreshResponse.ok) {
                    throw new FetchError(
                        refreshRes?.id || 'unknown',
                        "Session expired, please log in again",
                        {
                            status: refreshResponse.status,
                            isClientError: true,
                        }
                    );
                }
            } catch (error) {
                if (error instanceof FetchError) {
                    throw error;
                }
                throw new FetchError(
                    'network-error',
                    'Failed to refresh token due to network error',
                    { isNetworkError: true }
                );
            } finally {
                refreshTokenPromise = null;
            }
        })();
    }
    return refreshTokenPromise;
};

export const fetchClient = async <T>(
    url: string,
    options: FetchOptions = {}
): Promise<T> => {
    if (!url.startsWith("http")) {
        url = BASE_URL + url;
    }

    const defaultHeaders: Record<string, string> = {
        'Cache-Control': 'no-store',
        ...(options.method?.toUpperCase() !== "GET" && 
            options.method?.toUpperCase() !== "HEAD" && 
            { "Content-Type": "application/json" })
    };

    const headers = {
        ...defaultHeaders,
        ...(options.headers || {}),
    };

    try {
        const response = await fetch(url, {
            credentials: "include",
            ...options,
            headers,
        });

        if (!response.ok) {
            if (response.status === 401) {
                await refreshToken();
                
                // Retry the original request
                const retryResponse = await fetch(url, {
                    ...options,
                    credentials: "include",
                    headers,
                    body: options.body,
                });

                if (!retryResponse.ok) {
                    const retryRes = await retryResponse.json().catch(() => null) as ErrorResponse;
                    throw new FetchError(
                        retryRes?.id || 'unknown',
                        `HTTP error! Status: ${retryResponse.status}`,
                        {
                            status: retryResponse.status,
                            response: retryRes,
                        }
                    );
                }

                return validateAndParseResponse<T>(retryResponse);
            }

            const errorRes = await response.json().catch(() => null) as ErrorResponse;
            throw new FetchError(
                errorRes?.id || 'unknown',
                errorRes?.message || `HTTP error! Status: ${response.status}`,
                {
                    status: response.status,
                    response: errorRes,
                    isClientError: response.status >= 400 && response.status < 500,
                    isServerError: response.status >= 500,
                }
            );
        }

        return validateAndParseResponse<T>(response);
    } catch (error) {
        if (error instanceof TypeError) {
            throw new FetchError(
                'network-error',
                "Network error occurred, please try again",
                { isNetworkError: true }
            );
        }
        throw error;
    }
};

async function validateAndParseResponse<T>(response: Response): Promise<T> {
    const contentType = response.headers.get("content-type");

    if (response.status === 204 || !contentType || !contentType.includes("application/json")) {
        return {} as T;
    }

    try {
        const data = await response.json();
        return data as T;
    } catch (error) {
        throw new FetchError(
            'parse-error',
            'Failed to parse response as JSON',
            { isClientError: true }
        );
    }
}
