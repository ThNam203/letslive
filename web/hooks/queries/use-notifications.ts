import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { GetNotifications, GetUnreadCount } from "@/lib/api/notification";
import { unwrapResponse } from "@/lib/api/api-error";

export const NOTIFICATIONS_QUERY_KEY = ["notifications", "list"] as const;
export const NOTIFICATIONS_UNREAD_COUNT_QUERY_KEY = [
    "notifications",
    "unread-count",
] as const;

export function useNotificationsInfinite(enabled: boolean = true) {
    return useInfiniteQuery({
        queryKey: NOTIFICATIONS_QUERY_KEY,
        queryFn: async ({ pageParam }) =>
            unwrapResponse(await GetNotifications(pageParam)),
        initialPageParam: 0,
        getNextPageParam: (lastPage, allPages) =>
            lastPage.length > 0 ? allPages.length : undefined,
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
