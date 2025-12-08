export interface FetchOptions extends RequestInit {
    headers?: Record<string, string>;
    timeoutMs?: number;
    disableTimeout?: boolean;
}
