import { FetchError } from "../../types/fetch-error";
import { Livestream } from "../../types/livestream";
import { fetchClient } from "../../utils/fetchClient";

export async function GetVODInformation(vodId: string): Promise<{
    vod?: Livestream;
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<Livestream>(`/livestreams/${vodId}`);
        return { vod: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}

export async function GetLivestreamings(page: number = 0): Promise<{
    livestreamings: Livestream[];
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<Livestream[]>(`/livestreamings?page=${page}`);
        return { livestreamings: data };
    } catch (error) {
        return { livestreamings: [], fetchError: error as FetchError };
    }
}

export async function GetAllLivestreamOfUser(userId: string): Promise<{
    livestreams: Livestream[];
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<Livestream[]>(`/livestreams?userId=${userId}`);
        return { livestreams: data };
    } catch (error) {
        return { livestreams: [], fetchError: error as FetchError };
    }
}

export async function GetPopularVODs(page: number = 0): Promise<{
    vods: Livestream[];
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<Livestream[]>(`/popular-vods?page=${page}`);
        return { vods: data };
    } catch (error) {
        return { vods: [], fetchError: error as FetchError };
    }
}

export async function GetAllVODsAsAuthor(): Promise<{
    vods: Livestream[];
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<Livestream[]>(`/livestreams/author`);
        return { vods: data };
    } catch (error) {
        return { vods: [], fetchError: error as FetchError };
    }
}

export async function IsUserStreaming(userId: string): Promise<{
    isStreaming: boolean;
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<string>(`/is-streaming?userId=${userId}`);
        return { isStreaming: data === "true" };
    } catch (error) {
        return { isStreaming: false, fetchError: error as FetchError };
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
        await fetchClient<void>(`/livestreams/${vodId}`, { method: 'DELETE' });
        return { fetchError: undefined };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}