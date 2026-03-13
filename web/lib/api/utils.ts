import { ApiResponse } from "@/types/fetch-response";
import { fetchClient } from "@/utils/fetchClient";

export async function UploadFile(
    file: File,
): Promise<ApiResponse<{ newPath?: string }>> {
    const formData = new FormData();
    formData.append("file", file);

    return fetchClient<ApiResponse<{ newPath?: string }>>(`/upload-file`, {
        method: "POST",
        body: formData,
    });
}
