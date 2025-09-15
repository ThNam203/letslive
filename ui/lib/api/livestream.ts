import { FetchError } from "../../types/fetch-error";
import { Livestream } from "../../types/livestream";
import { fetchClient } from "@/utils/fetchClient";

export async function GetPopularLivestreams(page: number = 0, limit: number = 10): Promise<{
    livestreams: Livestream[];
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<Livestream[]>(`/popular-livestreams?page=${page}&limit=${limit}`);
        return { livestreams: data };
    } catch (error) {
        return { livestreams: [], fetchError: error as FetchError };
    }
}

export async function GetLivestreamOfUser(userId: string): Promise<{
    livestream: Livestream | null;
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<Livestream>(`/livestreams?userId=${userId}`);
        return { livestream: data };
    } catch (error) {
        return { livestream: null, fetchError: error as FetchError };
    }
}
