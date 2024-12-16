import { FetchError } from "@/types/fetch-error";
import { FetchOptions } from "@/types/fetch-options";
import { cache } from "react";

let isRefreshing = false;
let refreshSubscribers: Array<() => void> = [];
const BASE_URL = "http://localhost:8000";

const onTokenRefreshed = () => {
    refreshSubscribers.forEach((callback) => callback());
    refreshSubscribers = [];
};

const subscribeTokenRefresh = (callback: () => void) => {
    refreshSubscribers.push(callback);
};

export const fetchClient = async <T>(
    url: string,
    options: FetchOptions = {}
): Promise<T> => {
    if (!url.startsWith("http")) {
        url = BASE_URL + url;
    }

    const defaultHeaders: Record<string, string> = options.method?.toUpperCase() === "GET" || options.method?.toUpperCase() === "HEAD"
        ? { cache: "no-store" }
        : { cache: "no-store", "Content-Type": "application/json" };

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
                if (!isRefreshing) {
                    isRefreshing = true;

                    try {
                        const refreshResponse = await fetch(
                            `${BASE_URL}/auth/refresh-token`,
                            { method: "POST", credentials: "include" }
                        );

                        if (!refreshResponse.ok) {
                            refreshSubscribers = [];
                            throw new FetchError(
                                "Session expired, please log in again",
                                {
                                    status: 401,
                                    isClientError: true,
                                }
                            );
                        } else {
                            onTokenRefreshed(); // Notify waiting requests
                        }
                    } catch (refreshError) {
                        throw refreshError;
                    } finally {
                        isRefreshing = false;
                    }

                    // Wait for the token to be refreshed
                    return new Promise<T>((resolve, reject) => {
                        subscribeTokenRefresh(async () => {
                            try {
                                // Retry the original request
                                const retryResponse = await fetch(url, {
                                    ...options,
                                    credentials: "include",
                                });
                                if (!retryResponse.ok) {
                                    const retryErrorData = await retryResponse
                                        .json()
                                        .catch(() => null);
                                    throw new FetchError(
                                        `HTTP error! Status: ${retryResponse.status}`,
                                        {
                                            status: retryResponse.status,
                                            response: retryErrorData,
                                        }
                                    );
                                }
                                resolve((await retryResponse.json()) as T);
                            } catch (retryError) {
                                reject(retryError);
                            }
                        });
                    });
                }
            }

            const error: FetchError = new Error();
            error.status = response.status;
            error.response = await response.json().catch(() => null); // Safely parse JSON
            error.isClientError =
                response.status >= 399 && response.status < 500;
            error.isServerError = response.status >= 499;

            throw error;
        }

        return (await response.json()) as T;
    } catch (error: any) {
        if (error instanceof TypeError) {
            const networkError: FetchError = new Error(
                error.message || "Network request failed"
            );
            networkError.isNetworkError = true;
            throw networkError;
        }
        throw error;
    }
};