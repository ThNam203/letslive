import { ApiResponse } from "@/types/fetch-response";
import { Livestream } from "../../types/livestream";
import { fetchClient } from "@/utils/fetchClient";

export async function GetPopularLivestreams(
    page: number = 0,
    limit: number = 10,
): Promise<ApiResponse<Livestream[]>> {
    return fetchClient<ApiResponse<Livestream[]>>(
        `/popular-livestreams?page=${page}&limit=${limit}`,
    );
}

export async function GetLivestreamOfUser(
    userId: string,
): Promise<ApiResponse<Livestream | null>> {
    return fetchClient<ApiResponse<Livestream | null>>(
        `/livestreams?userId=${userId}`,
    );
}
