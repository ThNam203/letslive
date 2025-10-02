import { ApiResponse } from "@/types/fetch-response";
import { Livestream } from "../../types/livestream";
import { fetchClient } from "@/utils/fetchClient";

export async function GetPopularLivestreams(page: number = 0, limit: number = 10): Promise<ApiResponse<{ livestreams: Livestream[] }>> {
    return fetchClient<ApiResponse<{ livestreams: Livestream[] }>>(`/popular-livestreams?page=${page}&limit=${limit}`);
}

export async function GetLivestreamOfUser(userId: string): Promise<ApiResponse<{ livestream: Livestream | null }>> {
    return fetchClient<ApiResponse<{ livestream: Livestream | null }>>(`/livestreams?userId=${userId}`);
}
