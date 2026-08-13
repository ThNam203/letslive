import { InfiniteData, QueryClient } from "@tanstack/react-query";
import { VODComment } from "@/types/vod-comment";
import {
    CommentsPage,
    vodCommentsQueryKey,
} from "@/hooks/queries/use-vod-comments";

type CommentsData = InfiniteData<CommentsPage>;

// Total is only tracked on the first-fetched page (pages[0]) — that index
// never shifts as later pages get appended, so it's a stable place to read
// the running count from.
export function prependVodComment(
    queryClient: QueryClient,
    vodId: string,
    comment: VODComment,
) {
    queryClient.setQueryData<CommentsData>(vodCommentsQueryKey(vodId), (old) => {
        if (!old) return old;
        const [first, ...rest] = old.pages;
        return {
            ...old,
            pages: [
                {
                    items: [comment, ...(first?.items ?? [])],
                    total: (first?.total ?? 0) + 1,
                },
                ...rest,
            ],
        };
    });
}

export function markVodCommentDeleted(
    queryClient: QueryClient,
    vodId: string,
    commentId: string,
) {
    queryClient.setQueryData<CommentsData>(vodCommentsQueryKey(vodId), (old) => {
        if (!old) return old;
        const [first, ...rest] = old.pages;
        return {
            ...old,
            pages: [
                {
                    items: (first?.items ?? []).map((c) =>
                        c.id === commentId
                            ? { ...c, content: "", isDeleted: true }
                            : c,
                    ),
                    total: Math.max((first?.total ?? 0) - 1, 0),
                },
                ...rest.map((page) => ({
                    ...page,
                    items: page.items.map((c) =>
                        c.id === commentId
                            ? { ...c, content: "", isDeleted: true }
                            : c,
                    ),
                })),
            ],
        };
    });
}
