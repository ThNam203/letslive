import GLOBAL from "@/global";
import { ErrorResponse, FetchError } from "@/types/fetch-error";
import { FetchOptions } from "@/types/fetch-options";

// Singleton promise for token refresh
let refreshTokenPromise: Promise<string> | null = null;

const callRefreshToken = async (cookieString: string | null): Promise<string> => {
    if (!refreshTokenPromise) {
        refreshTokenPromise = (async () => {
            try {
                let fetchOptions: any = {
                    method: "POST",
                };

                if (cookieString) fetchOptions.headers = { "Cookie": cookieString };
                else fetchOptions.credentials = "include";

                const refreshResponse = await fetch(
                    `${GLOBAL.API_URL}/auth/refresh-token`,
                    fetchOptions
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

                return refreshResponse.headers.get("Set-Cookie")?.split(";")[0] + ";" || "";
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

const fetchClient = async <T>(
    url: string,
    options: FetchOptions = {}
): Promise<T> => {
    if (!url.startsWith("http")) {
        url = GLOBAL.API_URL + url;
    }

    const defaultHeaders: Record<string, string> = {
        'Cache-Control': 'no-store',
        ...(options.method?.toUpperCase() !== "GET" && 
            options.method?.toUpperCase() !== "HEAD" && 
            { "Content-Type": "application/json" })
    };

    const headers = {
        ...defaultHeaders,
        ...(options.headers ?? {}),
    };
    
    let includeCredentials = undefined;
    if (!options.headers || !options.headers.Cookie) includeCredentials = "include";

    try {
        const response = await fetch(url, {
            credentials: includeCredentials as RequestCredentials | undefined,
            ...options,
            headers,
        });

        if (!response.ok) {
            if (response.status === 401) {
                const newAccessToken = await callRefreshToken(options.headers && options.headers.Cookie ? options.headers.Cookie : null);

                if (includeCredentials === undefined) headers.Cookie = newAccessToken;
                
                // Retry the original request
                const retryResponse = await fetch(url, {
                    credentials: includeCredentials as RequestCredentials | undefined,
                    ...options,
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
