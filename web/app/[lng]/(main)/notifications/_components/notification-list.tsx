"use client";

import useT from "@/hooks/use-translation";
import { Button } from "@/components/ui/button";
import {
    NotificationCard,
    NotificationEmpty,
    NotificationLoading,
} from "@/components/notification";
import type { TimeAgoTranslator } from "@/components/notification";
import { Notification } from "@/types/notification";

type NotificationListProps = {
    notifications: Notification[];
    isLoading: boolean;
    hasMore: boolean;
    t: TimeAgoTranslator;
    onMarkAsRead: (id: string) => void;
    onDelete: (id: string) => void;
    onLoadMore: () => void;
};

export function NotificationList({
    notifications,
    isLoading,
    hasMore,
    t,
    onMarkAsRead,
    onDelete,
    onLoadMore,
}: NotificationListProps) {
    const { t: tNotif } = useT(["notification"]);
    const emptyClassName =
        "flex items-center justify-center py-20 text-sm text-muted-foreground";

    if (isLoading && notifications.length === 0) {
        return (
            <NotificationLoading
                message={tNotif("loading")}
                className={emptyClassName}
            />
        );
    }

    if (notifications.length === 0) {
        return (
            <NotificationEmpty
                message={tNotif("no_notifications_yet")}
                className={emptyClassName}
            />
        );
    }

    return (
        <div className="flex flex-col gap-2">
            {notifications.map((notification) => (
                <NotificationCard
                    key={notification.id}
                    notification={notification}
                    t={t}
                    onMarkAsRead={onMarkAsRead}
                    onDelete={onDelete}
                />
            ))}

            {hasMore && (
                <div className="flex justify-center py-4">
                    <Button
                        variant="outline"
                        className="cursor-pointer"
                        onClick={onLoadMore}
                        disabled={isLoading}
                    >
                        {isLoading ? tNotif("loading") : tNotif("load_more")}
                    </Button>
                </div>
            )}
        </div>
    );
}
