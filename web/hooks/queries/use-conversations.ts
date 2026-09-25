import { useInfiniteQuery } from "@tanstack/react-query";
import { GetConversations } from "@/lib/api/dm";
import { nextPageParam, unwrapPage } from "@/lib/query/paginated";

export const CONVERSATIONS_QUERY_KEY = ["dm", "conversations"] as const;
export const CONVERSATIONS_PAGE_SIZE = 20;

export function useConversationsInfinite(enabled: boolean) {
    return useInfiniteQuery({
        queryKey: CONVERSATIONS_QUERY_KEY,
        queryFn: async ({ pageParam }) =>
            unwrapPage(
                await GetConversations(pageParam, CONVERSATIONS_PAGE_SIZE),
            ),
        initialPageParam: 0,
        getNextPageParam: (_lastPage, allPages) =>
            nextPageParam(allPages, CONVERSATIONS_PAGE_SIZE),
        enabled,
    });
}
