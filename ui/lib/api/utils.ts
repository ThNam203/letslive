import { FetchError } from "../../types/fetch-error";
import { fetchClient } from "../../utils/fetchClient";

export async function UploadFile(
    file: File
): Promise<{ newPath?: string; fetchError?: FetchError }> {
    try {
        const formData = new FormData();
        formData.append("file", file);

        const data = await fetchClient<string>(`/upload-file`, {
            method: "POST",
            body: formData
        });

        return { newPath: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}