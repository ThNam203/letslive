import { useInfiniteQuery } from "@tanstack/react-query";
import { GetConversations } from "@/lib/api/dm";
import { unwrapResponse } from "@/lib/api/api-error";

export const CONVERSATIONS_QUERY_KEY = ["dm", "conversations"] as const;
const CONVERSATIONS_PAGE_SIZE = 20;

export function useConversationsInfinite(enabled: boolean) {
    return useInfiniteQuery({
        queryKey: CONVERSATIONS_QUERY_KEY,
        queryFn: async ({ pageParam }) =>
            unwrapResponse(
                await GetConversations(pageParam, CONVERSATIONS_PAGE_SIZE),
            ),
        initialPageParam: 0,
        getNextPageParam: (lastPage, allPages) =>
            lastPage.length > 0 ? allPages.length : undefined,
        enabled,
    });
}
