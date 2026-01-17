import { ApiResponse } from "@/types/fetch-response";
import { VOD } from "@/types/vod";
import { fetchClient } from "@/utils/fetchClient";

export async function GetAllVODsAsAuthor(): Promise<ApiResponse<VOD[]>> {
    return fetchClient<ApiResponse<VOD[]>>(`/vods/author`);
}

export async function GetPublicVODsOfUser(
    userId: string,
    page: number = 0,
    limit: number = 10,
): Promise<ApiResponse<VOD[]>> {
    return fetchClient<ApiResponse<VOD[]>>(
        `/vods?userId=${userId}&page=${page}&limit=${limit}`,
    );
}

export async function GetPopularVODs(
    page: number = 0,
    limit: number = 10,
): Promise<ApiResponse<VOD[]>> {
    return fetchClient<ApiResponse<VOD[]>>(
        `/popular-vods?page=${page}&limit=${limit}`,
    );
}

export async function GetVODInformation(
    vodId: string,
): Promise<ApiResponse<VOD | null>> {
    return fetchClient<ApiResponse<VOD | null>>(`/vods/${vodId}`);
}

export async function UpdateVOD(
    vodId: string,
    title: string,
    description: string,
    visibility: string,
    newThumbnail?: string,
): Promise<ApiResponse<void>> {
    const updateData = {
        title,
        description,
        visibility,
    } as any;

    if (newThumbnail) {
        updateData["thumbnailURL"] = newThumbnail;
    }

    return fetchClient<ApiResponse<void>>(`/vods/${vodId}`, {
        method: "PATCH",
        body: JSON.stringify(updateData),
    });
}

export async function DeleteVOD(vodId: string): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>(`/vods/${vodId}`, {
        method: "DELETE",
    });
}
