import { ApiError } from "@/lib/api/api-error";
import { ApiResponse } from "@/types/fetch-response";

/**
 * One page of a list endpoint, keeping the server-reported total that
 * `unwrapResponse` would otherwise discard.
 *
 * `total` is undefined when the endpoint sends no pagination meta (Go omits
 * `total` when it is 0, so 0 also arrives as undefined).
 */
export type PaginatedPage<T> = { items: T[]; total?: number };

export function unwrapPage<T>(res: ApiResponse<T[]>): PaginatedPage<T> {
    if (!res.success) throw new ApiError(res);
    return { items: res.data ?? [], total: res.meta?.total };
}

/**
 * Next page index for `useInfiniteQuery`, or undefined when the list is
 * exhausted. Uses `meta.total` when the endpoint reports it; otherwise falls
 * back to "a short page means the end".
 */
export function nextPageParam<T>(
    allPages: PaginatedPage<T>[],
    pageSize: number,
): number | undefined {
    const lastPage = allPages[allPages.length - 1];
    if (!lastPage) return undefined;

    const loaded = allPages.reduce((sum, page) => sum + page.items.length, 0);

    if (lastPage.total !== undefined) {
        return loaded < lastPage.total ? allPages.length : undefined;
    }

    return lastPage.items.length === pageSize ? allPages.length : undefined;
}

export function flattenPages<T>(
    data: { pages: PaginatedPage<T>[] } | undefined,
): T[] {
    return data?.pages.flatMap((page) => page.items) ?? [];
}
