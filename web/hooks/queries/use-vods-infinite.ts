import { useInfiniteQuery } from "@tanstack/react-query";
import { GetPopularVODs } from "@/lib/api/vod";
import { nextPageParam, unwrapPage } from "@/lib/query/paginated";

const VODS_PAGE_SIZE = 20;

export function useVodsInfinite() {
    return useInfiniteQuery({
        queryKey: ["vods-feed"],
        queryFn: async ({ pageParam }) =>
            unwrapPage(await GetPopularVODs(pageParam, VODS_PAGE_SIZE)),
        initialPageParam: 0,
        getNextPageParam: (_lastPage, allPages) =>
            nextPageParam(allPages, VODS_PAGE_SIZE),
    });
}
