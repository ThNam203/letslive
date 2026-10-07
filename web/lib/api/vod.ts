import { ApiResponse } from "@/types/fetch-response";
import { VOD, VODReaction, VODReactionState } from "@/types/vod";
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
    limit: number = 20,
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
        ...(newThumbnail ? { thumbnailUrl: newThumbnail } : {}),
    };

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

export async function RegisterVODView(
    vodId: string,
    watchedSeconds: number,
): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>(`/vods/${vodId}/view`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({ watchedSeconds }),
    });
}

export async function GetMyVODReaction(
    vodId: string,
): Promise<ApiResponse<VODReactionState>> {
    return fetchClient<ApiResponse<VODReactionState>>(
        `/vods/${vodId}/reaction`,
    );
}

export async function SetVODReaction(
    vodId: string,
    reaction: VODReaction,
): Promise<ApiResponse<VODReactionState>> {
    return fetchClient<ApiResponse<VODReactionState>>(
        `/vods/${vodId}/reaction`,
        {
            method: "PUT",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ reaction }),
        },
    );
}

export async function RemoveVODReaction(
    vodId: string,
): Promise<ApiResponse<VODReactionState>> {
    return fetchClient<ApiResponse<VODReactionState>>(
        `/vods/${vodId}/reaction`,
        {
            method: "DELETE",
        },
    );
}

export async function UploadVOD(
    file: File,
    title: string,
    description: string,
    visibility: string,
): Promise<ApiResponse<VOD>> {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("title", title);
    formData.append("description", description);
    formData.append("visibility", visibility);

    return fetchClient<ApiResponse<VOD>>(`/vods/upload`, {
        method: "POST",
        body: formData,
        disableTimeout: true,
    });
}
