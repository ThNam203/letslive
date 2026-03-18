import { ApiResponse } from "@/types/fetch-response";
import GLOBAL from "../global";

export type UploadProgressCallback = (loaded: number, total: number) => void;

/** Error codes for client-side upload failures; used by UI for i18n. */
export const UPLOAD_ERROR_CODES = {
    PARSE_RESPONSE: "manager_parse_error",
    FAILED: "manager_failed",
    CANCELLED: "manager_cancelled",
} as const;

export class UploadClientError extends Error {
    constructor(
        message: string,
        public readonly code: (typeof UPLOAD_ERROR_CODES)[keyof typeof UPLOAD_ERROR_CODES],
    ) {
        super(message);
        this.name = "UploadClientError";
    }
}

/**
 * Upload a file with progress tracking using XMLHttpRequest.
 * Returns a promise that resolves with the parsed API response,
 * and an abort function to cancel the upload.
 */
export function uploadWithProgress<T = ApiResponse<unknown>>(
    path: string,
    formData: FormData,
    onProgress?: UploadProgressCallback,
): { promise: Promise<T>; abort: () => void } {
    const url = path.startsWith("http") ? path : GLOBAL.API_URL + path;
    const xhr = new XMLHttpRequest();

    const promise = new Promise<T>((resolve, reject) => {
        xhr.open("POST", url);
        xhr.withCredentials = true;
        xhr.setRequestHeader("Cache-Control", "no-store");

        if (onProgress) {
            xhr.upload.addEventListener("progress", (e) => {
                if (e.lengthComputable) {
                    onProgress(e.loaded, e.total);
                }
            });
        }

        xhr.addEventListener("load", () => {
            try {
                const data = JSON.parse(xhr.responseText);
                resolve(data as T);
            } catch {
                reject(
                    new UploadClientError(
                        "Failed to parse response",
                        UPLOAD_ERROR_CODES.PARSE_RESPONSE,
                    ),
                );
            }
        });

        xhr.addEventListener("error", () => {
            reject(
                new UploadClientError(
                    "Upload failed",
                    UPLOAD_ERROR_CODES.FAILED,
                ),
            );
        });

        xhr.addEventListener("abort", () => {
            reject(
                new UploadClientError(
                    "Upload cancelled",
                    UPLOAD_ERROR_CODES.CANCELLED,
                ),
            );
        });

        xhr.send(formData);
    });

    return { promise, abort: () => xhr.abort() };
}
