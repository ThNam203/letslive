import { useInfiniteQuery } from "@tanstack/react-query";
import { GetVODComments } from "@/lib/api/vod-comment";
import { ApiError } from "@/lib/api/api-error";
import { VODComment } from "@/types/vod-comment";

const COMMENTS_PAGE_SIZE = 10;

export type CommentsPage = { items: VODComment[]; total: number };

export function vodCommentsQueryKey(vodId: string) {
    return ["vod-comments", vodId] as const;
}

export function useVodCommentsInfinite(vodId: string) {
    return useInfiniteQuery({
        queryKey: vodCommentsQueryKey(vodId),
        queryFn: async ({ pageParam }): Promise<CommentsPage> => {
            const res = await GetVODComments(vodId, pageParam, COMMENTS_PAGE_SIZE);
            if (!res.success) throw new ApiError(res);
            return { items: res.data ?? [], total: res.meta?.total ?? 0 };
        },
        initialPageParam: 0,
        getNextPageParam: (lastPage, allPages) =>
            lastPage.items.length === COMMENTS_PAGE_SIZE
                ? allPages.length
                : undefined,
    });
}
