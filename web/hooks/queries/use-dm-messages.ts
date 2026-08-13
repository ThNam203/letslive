import { useInfiniteQuery } from "@tanstack/react-query";
import { GetDmMessages } from "@/lib/api/dm";
import { unwrapResponse } from "@/lib/api/api-error";

const MESSAGES_PAGE_SIZE = 50;

export function dmMessagesQueryKey(conversationId: string) {
    return ["dm", "messages", conversationId] as const;
}

// Pages are ordered newest-batch-first; each page's own contents are
// oldest-to-newest. Flatten with `[...pages].reverse().flat()` to render
// chronologically. New live messages append onto the front page (pages[0]).
export function useDmMessagesInfinite(conversationId: string, enabled: boolean) {
    return useInfiniteQuery({
        queryKey: dmMessagesQueryKey(conversationId),
        queryFn: async ({ pageParam }) =>
            unwrapResponse(await GetDmMessages(conversationId, pageParam)),
        initialPageParam: undefined as string | undefined,
        getNextPageParam: (lastPage) =>
            lastPage.length >= MESSAGES_PAGE_SIZE ? lastPage[0]?._id : undefined,
        enabled,
    });
}
