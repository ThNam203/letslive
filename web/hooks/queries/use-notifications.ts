import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { GetNotifications, GetUnreadCount } from "@/lib/api/notification";
import { unwrapResponse } from "@/lib/api/api-error";
import { nextPageParam, unwrapPage } from "@/lib/query/paginated";

export const NOTIFICATIONS_QUERY_KEY = ["notifications", "list"] as const;
export const NOTIFICATIONS_UNREAD_COUNT_QUERY_KEY = [
    "notifications",
    "unread-count",
] as const;

// the endpoint takes no limit: the backend pages notifications 20 at a time
const NOTIFICATIONS_PAGE_SIZE = 20;

export function useNotificationsInfinite(enabled: boolean = true) {
    return useInfiniteQuery({
        queryKey: NOTIFICATIONS_QUERY_KEY,
        queryFn: async ({ pageParam }) =>
            unwrapPage(await GetNotifications(pageParam)),
        initialPageParam: 0,
        getNextPageParam: (_lastPage, allPages) =>
            nextPageParam(allPages, NOTIFICATIONS_PAGE_SIZE),
        enabled,
    });
}

export function useUnreadNotificationCount(enabled: boolean) {
    return useQuery({
        queryKey: NOTIFICATIONS_UNREAD_COUNT_QUERY_KEY,
        queryFn: async () => unwrapResponse(await GetUnreadCount()),
        enabled,
        refetchInterval: 30_000,
        refetchIntervalInBackground: false,
    });
}
