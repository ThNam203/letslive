import { VOD } from "@/types/vod";
import { FetchError } from "../../types/fetch-error";
import { fetchClient } from "@/utils/fetchClient";

export async function GetAllVODsAsAuthor(): Promise<{
    vods: VOD[];
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<VOD[]>(`/vods/author`);
        return { vods: data };
    } catch (error) {
        return { vods: [], fetchError: error as FetchError };
    }
}

export async function GetPublicVODsOfUser(userId: string, page: number = 0, limit: number = 10): Promise<{
    vods: VOD[];
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<VOD[]>(`/vods?userId=${userId}&page=${page}&limit=${limit}`);
        return { vods: data };
    } catch (error) {
        return { vods: [], fetchError: error as FetchError };
    }
}

export async function GetPopularVODs(page: number = 0, limit: number = 10): Promise<{
    vods: VOD[];
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<VOD[]>(`/popular-vods?page=${page}&limit=${limit}`);
        return { vods: data };
    } catch (error) {
        return { vods: [], fetchError: error as FetchError };
    }
}
export async function GetVODInformation(vodId: string): Promise<{
    vod?: VOD;
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<VOD>(`/vods/${vodId}`);
        return { vod: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}


export async function UpdateVOD(vodId: string, title: string, description: string, visibility: string, newThumbnail?: string): Promise<{
    fetchError?: FetchError;
}> {
    const updateData = {
        title,
        description,
        visibility,
    } as any

    if (newThumbnail) {
        updateData['thumbnailURL'] = newThumbnail;
    }

    try {
        await fetchClient<void>(`/livestreams/${vodId}`, { method: 'PATCH', body: JSON.stringify(updateData) });
        return { fetchError: undefined };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}

export async function DeleteVOD(vodId: string): Promise<{
    fetchError?: FetchError;
}> {
    try {
        await fetchClient<void>(`/vods/${vodId}`, { method: 'DELETE' });
        return { fetchError: undefined };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}