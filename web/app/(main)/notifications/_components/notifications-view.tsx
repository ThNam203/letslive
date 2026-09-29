"use client";

import useT from "@/hooks/use-translation";
import { NotificationPageHeader } from "./notification-page-header";
import { NotificationList } from "./notification-list";
import { useNotificationsInfinite } from "@/hooks/queries/use-notifications";
import { flattenPages } from "@/lib/query/paginated";
import {
    useDeleteNotification,
    useMarkAllNotificationsAsRead,
    useMarkNotificationAsRead,
} from "@/hooks/queries/use-notification-mutations";

export default function NotificationsView({
    heading,
}: {
    heading: React.ReactNode;
}) {
    const { t } = useT(["notification", "common"]);
    const { data, isLoading, hasNextPage, fetchNextPage, isFetchingNextPage } =
        useNotificationsInfinite();
    const notifications = flattenPages(data);

    const markAsRead = useMarkNotificationAsRead();
    const markAllAsRead = useMarkAllNotificationsAsRead();
    const deleteNotification = useDeleteNotification();

    return (
        <>
            <NotificationPageHeader
                heading={heading}
                hasUnread={notifications.some((n) => !n.isRead)}
                onMarkAllAsRead={() => markAllAsRead.mutate()}
            />

            <NotificationList
                notifications={notifications}
                isLoading={isLoading || isFetchingNextPage}
                hasMore={!!hasNextPage}
                t={t}
                onMarkAsRead={(id) => markAsRead.mutate(id)}
                onDelete={(id) => deleteNotification.mutate(id)}
                onLoadMore={() => fetchNextPage()}
            />
        </>
    );
}
