import GLOBAL from "../global";
import { ErrorResponse, FetchError } from "../types/fetch-error";
import { FetchOptions } from "../types/fetch-options";

// Singleton promise for token refresh
let refreshTokenPromise: Promise<void> | null = null;

const refreshToken = async (): Promise<void> => {
    if (!refreshTokenPromise) {
        refreshTokenPromise = (async () => {
            try {
                const refreshResponse = await fetch(
                    `${GLOBAL.API_URL}/auth/refresh-token`,
                    { method: "POST", credentials: "include" }
                );

                const refreshRes = await refreshResponse.json().catch(() => null);
                if (!refreshResponse.ok) {
                    throw new FetchError(
                        refreshRes?.id || 'unknown',
                        refreshRes.message || `Failed to refresh token! Status: ${refreshResponse.status}`,
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

// the T should be of type Object or string only
export const fetchClient = async <T>(
    url: string,
    options: FetchOptions = {}
): Promise<T> => {
    if (!url.startsWith("http")) {
        url = GLOBAL.API_URL + url;
    }

    const defaultHeaders: Record<string, string> = {
        'Cache-Control': 'no-store',
    };

    const headers = {
        ...defaultHeaders,
        ...(options.headers ?? {}),
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
                        retryRes.message,
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
                errorRes?.message ?? `Something wrong happened!`,
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
        console.log("Error in fetchClient:", error);
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

    if (response.status === 204 || !contentType) {
        return {} as T;
    }

    try {
        if (contentType.includes("text/plain")) {
            return (await response.text()) as T;
        }

        const data = await response.json();
        return data as T;
    } catch (error) {
        throw new FetchError(
            'parse-error',
            'Failed to parse response',
            { isClientError: true }
        );
    }
}