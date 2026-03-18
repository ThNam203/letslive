import { ApiResponse } from "@/types/fetch-response";
import GLOBAL from "../global";

export type UploadProgressCallback = (loaded: number, total: number) => void;

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
                reject(new Error("Failed to parse response"));
            }
        });

        xhr.addEventListener("error", () => {
            reject(new Error("Upload failed"));
        });

        xhr.addEventListener("abort", () => {
            reject(new Error("Upload cancelled"));
        });

        xhr.send(formData);
    });

    return { promise, abort: () => xhr.abort() };
}
