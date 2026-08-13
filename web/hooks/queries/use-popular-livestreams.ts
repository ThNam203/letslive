import { useQuery } from "@tanstack/react-query";
import { GetPopularLivestreams } from "@/lib/api/livestream";
import { unwrapResponse } from "@/lib/api/api-error";

export function usePopularLivestreams(page: number = 0, limit: number = 10) {
    return useQuery({
        queryKey: ["popular-livestreams", page, limit],
        queryFn: async () =>
            unwrapResponse(await GetPopularLivestreams(page, limit)),
    });
}
