"use client";

import { useCallback, useEffect, useState } from "react";
import useUser from "@/hooks/user";
import useNotification from "@/hooks/notification";
import {
    GetNotifications,
    MarkNotificationAsRead,
    MarkAllNotificationsAsRead,
    DeleteNotification,
} from "@/lib/api/notification";
import { Notification } from "@/types/notification";

export function useNotificationsPage() {
    const userState = useUser();
    const [notifications, setNotifications] = useState<Notification[]>([]);
    const [page, setPage] = useState(0);
    const [total, setTotal] = useState(0);
    const [isLoading, setIsLoading] = useState(true);

    const fetchNotifications = useCallback(
        async (pageNum: number) => {
            if (!userState.user) return;
            setIsLoading(true);
            try {
                const res = await GetNotifications(pageNum);
                if (res.success && res.data) {
                    if (pageNum === 0) {
                        setNotifications(res.data);
                    } else {
                        setNotifications((prev) => [...prev, ...res.data!]);
                    }
                    if (res.meta) {
                        setTotal(res.meta.total ?? 0);
                    }
                }
            } catch {
                // silently ignore
            } finally {
                setIsLoading(false);
            }
        },
        [userState.user],
    );

    useEffect(() => {
        fetchNotifications(0);
    }, [fetchNotifications]);

    const handleMarkAsRead = useCallback(
        async (notificationId: string) => {
            try {
                const res = await MarkNotificationAsRead(notificationId);
                if (res.success) {
                    setNotifications((prev) =>
                        prev.map((n) =>
                            n.id === notificationId
                                ? { ...n, isRead: true }
                                : n,
                        ),
                    );
                    useNotification.getState().markAsRead(notificationId);
                }
            } catch {
                // silently ignore
            }
        },
        [],
    );

    const handleMarkAllAsRead = useCallback(async () => {
        try {
            const res = await MarkAllNotificationsAsRead();
            if (res.success) {
                setNotifications((prev) =>
                    prev.map((n) => ({ ...n, isRead: true })),
                );
                useNotification.getState().markAllAsRead();
            }
        } catch {
            // silently ignore
        }
    }, []);

    const handleDelete = useCallback(
        async (notificationId: string) => {
            try {
                const res = await DeleteNotification(notificationId);
                if (res.success) {
                    setNotifications((prev) =>
                        prev.filter((n) => n.id !== notificationId),
                    );
                    useNotification.getState().removeNotification(notificationId);
                    setTotal((prev) => prev - 1);
                }
            } catch {
                // silently ignore
            }
        },
        [],
    );

    const handleLoadMore = useCallback(() => {
        const nextPage = page + 1;
        setPage(nextPage);
        fetchNotifications(nextPage);
    }, [page, fetchNotifications]);

    const canAccess = !!userState.user;
    const hasMore = notifications.length < total;

    return {
        notifications,
        isLoading,
        page,
        total,
        canAccess,
        hasMore,
        fetchNotifications,
        handleMarkAsRead,
        handleMarkAllAsRead,
        handleDelete,
        handleLoadMore,
    };
}
