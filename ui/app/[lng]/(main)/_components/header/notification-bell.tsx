"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useParams } from "next/navigation";
import useUser from "@/hooks/user";
import useNotification from "@/hooks/notification";
import useT from "@/hooks/use-translation";
import {
    GetNotifications,
    GetUnreadCount,
    MarkNotificationAsRead,
    MarkAllNotificationsAsRead,
} from "@/lib/api/notification";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover";
import IconBell from "@/components/icons/bell";
import {
    NotificationPopupContent,
    NOTIFICATION_POLL_INTERVAL_MS,
} from "@/components/notification";

export default function NotificationBell() {
    const { t } = useT(["notification"]);
    const params = useParams();
    const lng = (params?.lng as string) ?? "en";
    const userState = useUser();
    const notifState = useNotification();
    const [isOpen, setIsOpen] = useState(false);
    const lastFetchRef = useRef<number>(0);
    const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

    const fetchUnreadCount = useCallback(async () => {
        if (!userState.user) return;
        try {
            const res = await GetUnreadCount();
            if (res.success && res.data) {
                useNotification.getState().setUnreadCount(res.data.count);
            }
        } catch {
            // silently ignore
        }
    }, [userState.user]);

    useEffect(() => {
        if (!userState.user) return;
        fetchUnreadCount();
        intervalRef.current = setInterval(() => {
            if (!document.hidden) fetchUnreadCount();
        }, NOTIFICATION_POLL_INTERVAL_MS);
        return () => {
            if (intervalRef.current) clearInterval(intervalRef.current);
        };
    }, [userState.user, fetchUnreadCount]);

    const canShowNotifications = !!userState.user;

    const handleOpenChange = useCallback(
        async (open: boolean) => {
            setIsOpen(open);
            if (
                open &&
                Date.now() - lastFetchRef.current >= NOTIFICATION_POLL_INTERVAL_MS
            ) {
                useNotification.getState().setIsLoading(true);
                try {
                    const res = await GetNotifications(0);
                    if (res.success && res.data) {
                        useNotification.getState().setNotifications(res.data);
                    }
                } catch {
                    // silently ignore
                } finally {
                    useNotification.getState().setIsLoading(false);
                    lastFetchRef.current = Date.now();
                }
            }
        },
        [],
    );

    const handleMarkAsRead = useCallback(
        async (notificationId: string) => {
            try {
                const res = await MarkNotificationAsRead(notificationId);
                if (res.success)
                    useNotification.getState().markAsRead(notificationId);
            } catch {
                // silently ignore
            }
        },
        [],
    );

    const handleMarkAllAsRead = useCallback(async () => {
        try {
            const res = await MarkAllNotificationsAsRead();
            if (res.success) useNotification.getState().markAllAsRead();
        } catch {
            // silently ignore
        }
    }, []);

    const handleNotificationClick = useCallback(
        (notificationId: string, isRead: boolean) => {
            if (!isRead) handleMarkAsRead(notificationId);
            setIsOpen(false);
        },
        [handleMarkAsRead],
    );

    if (!canShowNotifications) return null;

    return (
        <Popover open={isOpen} onOpenChange={handleOpenChange}>
            <PopoverTrigger asChild>
                <button className="relative cursor-pointer rounded-md p-1.5 transition-colors hover:bg-muted">
                    <IconBell className="size-5" />
                    {notifState.unreadCount > 0 && (
                        <span className="absolute -top-0.5 -right-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-destructive px-1 text-[10px] font-bold text-white">
                            {notifState.unreadCount > 99
                                ? "99+"
                                : notifState.unreadCount}
                        </span>
                    )}
                </button>
            </PopoverTrigger>
            <PopoverContent
                className="mr-4 w-80 border-border bg-muted p-0"
                align="end"
            >
                <NotificationPopupContent
                    notifications={notifState.notifications}
                    isLoading={notifState.isLoading}
                    unreadCount={notifState.unreadCount}
                    viewAllHref={`/${lng}/notifications`}
                    t={t}
                    onMarkAllAsRead={handleMarkAllAsRead}
                    onNotificationClick={handleNotificationClick}
                    onClose={() => setIsOpen(false)}
                />
            </PopoverContent>
        </Popover>
    );
}
