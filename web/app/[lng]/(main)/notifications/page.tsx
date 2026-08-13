"use client";

import useT from "@/hooks/use-translation";
import { useNotificationsPage } from "@/hooks/use-notifications-page";
import { NotificationPageHeader } from "./_components/notification-page-header";
import { NotificationList } from "./_components/notification-list";
import RequireAuth from "@/components/wrappers/RequireAuth";

export default function NotificationsPage() {
    const { t } = useT(["notification", "common"]);
    const {
        notifications,
        isLoading,
        hasMore,
        handleMarkAsRead,
        handleMarkAllAsRead,
        handleDelete,
        handleLoadMore,
    } = useNotificationsPage();

    return (
        <RequireAuth>
            <div className="small-scrollbar h-full min-h-0 overflow-auto">
                <div className="mx-auto w-full px-4 py-6">
                    <NotificationPageHeader
                        hasUnread={notifications.some((n) => !n.isRead)}
                        onMarkAllAsRead={handleMarkAllAsRead}
                    />

                    <NotificationList
                        notifications={notifications}
                        isLoading={isLoading}
                        hasMore={hasMore}
                        t={t}
                        onMarkAsRead={handleMarkAsRead}
                        onDelete={handleDelete}
                        onLoadMore={handleLoadMore}
                    />
                </div>
            </div>
        </RequireAuth>
    );
}
