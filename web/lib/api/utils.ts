import { ApiResponse } from "@/types/fetch-response";
import { fetchClient } from "@/utils/fetchClient";

export async function UploadFile(
    file: File,
): Promise<ApiResponse<{ newPath?: string }>> {
    const formData = new FormData();
    formData.append("file", file);

    const response = await fetchClient<
        ApiResponse<{ newPath?: string } | string>
    >(`/upload-file`, {
        method: "POST",
        body: formData,
    });

    if (response.success && typeof response.data === "string") {
        return {
            ...response,
            data: {
                newPath: response.data,
            },
        };
    }

    return response as ApiResponse<{ newPath?: string }>;
}
