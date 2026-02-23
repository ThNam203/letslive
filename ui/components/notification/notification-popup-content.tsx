"use client";

import Link from "next/link";
import useT from "@/hooks/use-translation";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { NotificationItem } from "./notification-item";
import { NotificationEmpty } from "./notification-empty";
import { NotificationLoading } from "./notification-loading";
import { Notification } from "@/types/notification";
import { TimeAgoTranslator } from "./utils";
import { NOTIFICATION_POPUP_LIMIT } from "./utils";

type NotificationPopupContentProps = {
    notifications: Notification[];
    isLoading: boolean;
    unreadCount: number;
    viewAllHref: string;
    t: TimeAgoTranslator;
    onMarkAllAsRead: () => void;
    onNotificationClick: (notificationId: string, isRead: boolean) => void;
    onClose: () => void;
};

export function NotificationPopupContent({
    notifications,
    isLoading,
    unreadCount,
    viewAllHref,
    t,
    onMarkAllAsRead,
    onNotificationClick,
    onClose,
}: NotificationPopupContentProps) {
    const { t: tNotif } = useT(["notification"]);
    return (
        <>
            <div className="flex items-center justify-between border-b border-border px-4 py-3">
                <h3 className="text-sm font-semibold text-foreground">
                    {tNotif("title")}
                </h3>
                {unreadCount > 0 && (
                    <Button
                        variant="ghost"
                        className="h-auto px-2 py-1 text-xs"
                        onClick={onMarkAllAsRead}
                    >
                        {tNotif("mark_all_as_read")}
                    </Button>
                )}
            </div>

            <ScrollArea className="h-80">
                {isLoading ? (
                    <NotificationLoading message={tNotif("loading")} />
                ) : notifications.length === 0 ? (
                    <NotificationEmpty message={tNotif("no_notifications")} />
                ) : (
                    <div className="flex flex-col">
                        {notifications.slice(0, NOTIFICATION_POPUP_LIMIT).map(
                            (notification) => (
                                <NotificationItem
                                    key={notification.id}
                                    notification={notification}
                                    t={t}
                                    variant="compact"
                                    onClick={() =>
                                        onNotificationClick(
                                            notification.id,
                                            notification.isRead,
                                        )
                                    }
                                />
                            ),
                        )}
                    </div>
                )}
            </ScrollArea>

            <div className="border-t border-border px-4 py-2">
                <Link
                    href={viewAllHref}
                    className="block text-center text-xs font-medium text-primary hover:underline"
                    onClick={onClose}
                >
                    {tNotif("view_all")}
                </Link>
            </div>
        </>
    );
}
