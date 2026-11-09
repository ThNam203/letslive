import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { SearchUsersByUsername } from "@/lib/api/user";
import { SearchVODs } from "@/lib/api/vod";
import { SearchLivestreams } from "@/lib/api/livestream";
import { unwrapResponse } from "@/lib/api/api-error";
import { nextPageParam, unwrapPage } from "@/lib/query/paginated";
import { ApiResponse } from "@/types/fetch-response";
import { PublicUser } from "@/types/user";
import { Livestream } from "@/types/livestream";
import { VOD } from "@/types/vod";
import { SearchType } from "@/lib/search";

export type SearchItem = { users: PublicUser; lives: Livestream; vods: VOD };

export const SEARCH_PREVIEW_LIMIT = 5;
const SEARCH_PAGE_SIZE = 20;

// Typing back to a previous term is common, so results are kept briefly
// rather than re-requested on every keystroke that retraces one.
const SEARCH_STALE_TIME_MS = 30_000;

const searchFns: {
    [T in SearchType]: (
        query: string,
        page: number,
        limit: number,
    ) => Promise<ApiResponse<SearchItem[T][]>>;
} = {
    users: SearchUsersByUsername,
    lives: SearchLivestreams,
    vods: SearchVODs,
};

function searchQueryKey(type: SearchType, query: string) {
    return ["search", type, query] as const;
}

export function useSearchPreview<T extends SearchType>(type: T, query: string) {
    const trimmed = query.trim();
    const search = searchFns[type];
    return useQuery({
        queryKey: [...searchQueryKey(type, trimmed), "preview"],
        queryFn: async () =>
            unwrapResponse(await search(trimmed, 0, SEARCH_PREVIEW_LIMIT)),
        enabled: trimmed.length > 0,
        staleTime: SEARCH_STALE_TIME_MS,
    });
}

export function useSearchInfinite<T extends SearchType>(
    type: T,
    query: string,
) {
    const trimmed = query.trim();
    const search = searchFns[type];
    return useInfiniteQuery({
        queryKey: [...searchQueryKey(type, trimmed), "paged"],
        queryFn: async ({ pageParam }) =>
            unwrapPage(await search(trimmed, pageParam, SEARCH_PAGE_SIZE)),
        initialPageParam: 0,
        getNextPageParam: (_lastPage, allPages) =>
            nextPageParam(allPages, SEARCH_PAGE_SIZE),
        enabled: trimmed.length > 0,
        staleTime: SEARCH_STALE_TIME_MS,
    });
}
