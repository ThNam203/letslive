import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
    MarkNotificationAsRead,
    MarkAllNotificationsAsRead,
    DeleteNotification,
} from "@/lib/api/notification";
import { unwrapResponse } from "@/lib/api/api-error";
import {
    NOTIFICATIONS_QUERY_KEY,
    NOTIFICATIONS_UNREAD_COUNT_QUERY_KEY,
} from "./use-notifications";

function useInvalidateNotifications() {
    const queryClient = useQueryClient();
    return () => {
        queryClient.invalidateQueries({ queryKey: NOTIFICATIONS_QUERY_KEY });
        queryClient.invalidateQueries({
            queryKey: NOTIFICATIONS_UNREAD_COUNT_QUERY_KEY,
        });
    };
}

export function useMarkNotificationAsRead() {
    const invalidate = useInvalidateNotifications();
    return useMutation({
        mutationFn: async (notificationId: string) =>
            unwrapResponse(await MarkNotificationAsRead(notificationId)),
        onSuccess: invalidate,
    });
}

export function useMarkAllNotificationsAsRead() {
    const invalidate = useInvalidateNotifications();
    return useMutation({
        mutationFn: async () =>
            unwrapResponse(await MarkAllNotificationsAsRead()),
        onSuccess: invalidate,
    });
}

export function useDeleteNotification() {
    const invalidate = useInvalidateNotifications();
    return useMutation({
        mutationFn: async (notificationId: string) =>
            unwrapResponse(await DeleteNotification(notificationId)),
        onSuccess: invalidate,
    });
}
