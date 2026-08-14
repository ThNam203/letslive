import { useInfiniteQuery } from "@tanstack/react-query";
import { GetPopularVODs } from "@/lib/api/vod";
import { unwrapResponse } from "@/lib/api/api-error";

const VODS_PAGE_SIZE = 20;

export function useVodsInfinite() {
    return useInfiniteQuery({
        queryKey: ["vods-feed"],
        queryFn: async ({ pageParam }) =>
            unwrapResponse(await GetPopularVODs(pageParam, VODS_PAGE_SIZE)),
        initialPageParam: 0,
        getNextPageParam: (lastPage, allPages) =>
            lastPage.length > 0 ? allPages.length : undefined,
    });
}
