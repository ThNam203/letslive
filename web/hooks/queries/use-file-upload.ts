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
            const results = await Promise.all(
                files.map((file) => UploadFile(file)),
            );

            const uploadedUrls: string[] = [];
            let hasError = false;

            for (const res of results) {
                if (res.success && res.data?.newPath) {
                    uploadedUrls.push(res.data.newPath);
                } else {
                    hasError = true;
                }
            }

            return { uploadedUrls, hasError };
        },
    });
}
