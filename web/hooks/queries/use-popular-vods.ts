import { useQuery } from "@tanstack/react-query";
import { GetPopularVODs } from "@/lib/api/vod";
import { unwrapResponse } from "@/lib/api/api-error";

export function usePopularVods(page: number = 0, limit: number = 20) {
    return useQuery({
        queryKey: ["popular-vods", page, limit],
        queryFn: async () => unwrapResponse(await GetPopularVODs(page, limit)),
    });
}
