"use client";

import Link from "next/link";
import useT from "@/hooks/use-translation";
import { Button } from "@/components/ui/button";
import { NotificationItemContent } from "./notification-item";
import { Notification } from "@/types/notification";
import { TimeAgoTranslator } from "./utils";
import { cn } from "@/utils/cn";

type NotificationCardProps = {
    notification: Notification;
    t: TimeAgoTranslator;
    onMarkAsRead: (id: string) => void;
    onDelete: (id: string) => void;
};

export function NotificationCard({
    notification,
    t,
    onMarkAsRead,
    onDelete,
}: NotificationCardProps) {
    const { t: tNotif } = useT(["notification"]);
    return (
        <div
            className={cn(
                "flex flex-col gap-1 rounded-lg border border-border p-4 transition-colors",
                !notification.isRead && "bg-primary/5",
            )}
        >
            <NotificationItemContent
                notification={notification}
                t={t}
                variant="full"
            />
            <div className="mt-1 flex items-center gap-2 pl-4">
                {notification.actionUrl && (
                    <Link
                        href={notification.actionUrl}
                        className="cursor-pointer text-xs font-medium text-primary hover:underline"
                        onClick={() => {
                            if (!notification.isRead) {
                                onMarkAsRead(notification.id);
                            }
                        }}
                    >
                        {notification.actionLabel ?? tNotif("view")}
                    </Link>
                )}
                {!notification.isRead && (
                    <Button
                        variant="ghost"
                        className="cursor-pointer h-auto px-2 py-0.5 text-xs"
                        onClick={() => onMarkAsRead(notification.id)}
                    >
                        {tNotif("mark_as_read")}
                    </Button>
                )}
                <Button
                    variant="ghost"
                    className="cursor-pointer h-auto px-2 py-0.5 text-xs text-destructive hover:text-destructive"
                    onClick={() => onDelete(notification.id)}
                >
                    {tNotif("delete")}
                </Button>
            </div>
        </div>
    );
}
