import { useMutation } from "@tanstack/react-query";
import { UploadFile } from "@/lib/api/utils";

export type UploadFilesResult = {
    /** Paths of the files that made it, in no particular order. */
    uploadedUrls: string[];
    /** True when at least one file in the batch failed. */
    hasError: boolean;
};

/**
 * Uploads a batch of files and reports partial success instead of throwing:
 * a message with three attachments where one fails is still worth sending
 * with the other two, so the caller decides what to do about the gap.
 */
export function useUploadFiles() {
    return useMutation({
        mutationFn: async (files: File[]): Promise<UploadFilesResult> => {
            // allSettled, not all: one upload rejecting must not discard the
            // paths of the ones that already succeeded
            const results = await Promise.allSettled(
                files.map((file) => UploadFile(file)),
            );

            const uploadedUrls: string[] = [];
            let hasError = false;

            for (const result of results) {
                const path =
                    result.status === "fulfilled" && result.value.success
                        ? result.value.data?.newPath
                        : undefined;

                if (path) {
                    uploadedUrls.push(path);
                } else {
                    hasError = true;
                }
            }

            return { uploadedUrls, hasError };
        },
    });
}
