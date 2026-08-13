"use client";

import useT from "@/hooks/use-translation";
import { NotificationPageHeader } from "./_components/notification-page-header";
import { NotificationList } from "./_components/notification-list";
import RequireAuth from "@/components/wrappers/RequireAuth";
import { useNotificationsInfinite } from "@/hooks/queries/use-notifications";
import {
    useDeleteNotification,
    useMarkAllNotificationsAsRead,
    useMarkNotificationAsRead,
} from "@/hooks/queries/use-notification-mutations";

export default function NotificationsPage() {
    const { t } = useT(["notification", "common"]);
    const { data, isLoading, hasNextPage, fetchNextPage, isFetchingNextPage } =
        useNotificationsInfinite();
    const notifications = data?.pages.flat() ?? [];

    const markAsRead = useMarkNotificationAsRead();
    const markAllAsRead = useMarkAllNotificationsAsRead();
    const deleteNotification = useDeleteNotification();

    return (
        <RequireAuth>
            <div className="small-scrollbar h-full min-h-0 overflow-auto">
                <div className="mx-auto w-full px-4 py-6">
                    <NotificationPageHeader
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
                </div>
            </div>
        </RequireAuth>
    );
}
