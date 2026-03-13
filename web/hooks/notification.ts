import { create } from "zustand";
import { Notification } from "@/types/notification";

export type NotificationState = {
    unreadCount: number;
    notifications: Notification[];
    isLoading: boolean;
    setUnreadCount: (count: number) => void;
    setNotifications: (notifications: Notification[]) => void;
    appendNotifications: (notifications: Notification[]) => void;
    markAsRead: (notificationId: string) => void;
    markAllAsRead: () => void;
    removeNotification: (notificationId: string) => void;
    setIsLoading: (isLoading: boolean) => void;
};

const useNotification = create<NotificationState>((set) => ({
    unreadCount: 0,
    notifications: [],
    isLoading: false,

    setUnreadCount: (count) => set({ unreadCount: count }),
    setNotifications: (notifications) => set({ notifications }),
    appendNotifications: (newNotifications) =>
        set((prev) => ({
            notifications: [...prev.notifications, ...newNotifications],
        })),
    markAsRead: (notificationId) =>
        set((prev) => ({
            notifications: prev.notifications.map((n) =>
                n.id === notificationId ? { ...n, isRead: true } : n,
            ),
            unreadCount: Math.max(0, prev.unreadCount - 1),
        })),
    markAllAsRead: () =>
        set((prev) => ({
            notifications: prev.notifications.map((n) => ({
                ...n,
                isRead: true,
            })),
            unreadCount: 0,
        })),
    removeNotification: (notificationId) =>
        set((prev) => {
            const removed = prev.notifications.find(
                (n) => n.id === notificationId,
            );
            return {
                notifications: prev.notifications.filter(
                    (n) => n.id !== notificationId,
                ),
                unreadCount:
                    removed && !removed.isRead
                        ? Math.max(0, prev.unreadCount - 1)
                        : prev.unreadCount,
            };
        }),
    setIsLoading: (isLoading) => set({ isLoading }),
}));

export default useNotification;
