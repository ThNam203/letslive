import { useQuery } from "@tanstack/react-query";
import { GetFollowingChannels, GetRecommendedChannels } from "@/lib/api/user";
import { unwrapResponse } from "@/lib/api/api-error";

export function useFollowingChannels(enabled: boolean) {
    return useQuery({
        queryKey: ["following-channels"],
        queryFn: async () => unwrapResponse(await GetFollowingChannels()),
        enabled,
    });
}

export function useRecommendedChannels(page: number = 0) {
    return useQuery({
        queryKey: ["recommended-channels", page],
        queryFn: async () =>
            unwrapResponse(await GetRecommendedChannels(page)),
    });
}
