"use client";

import Link from "next/link";
import { useCallback, useState } from "react";
import { useParams } from "next/navigation";
import useUser from "@/hooks/user";
import useT from "@/hooks/use-translation";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover";
import IconBell from "@/components/icons/bell";
import { NotificationPopupContent } from "@/components/notification";
import { I18N_FALLBACK_LNG } from "@/lib/i18n/settings";
import {
    useNotificationsInfinite,
    useUnreadNotificationCount,
} from "@/hooks/queries/use-notifications";
import {
    useMarkAllNotificationsAsRead,
    useMarkNotificationAsRead,
} from "@/hooks/queries/use-notification-mutations";

export default function NotificationBell() {
    const { t } = useT(["notification"]);
    const params = useParams();
    const lng = (params?.lng as string) ?? I18N_FALLBACK_LNG;
    const user = useUser((state) => state.user);
    const [isOpen, setIsOpen] = useState(false);

    const { data: unreadData } = useUnreadNotificationCount(!!user);
    const unreadCount = unreadData?.count ?? 0;

    const { data, isLoading } = useNotificationsInfinite(!!user && isOpen);
    const notifications = data?.pages[0] ?? [];

    const markAsRead = useMarkNotificationAsRead();
    const markAllAsRead = useMarkAllNotificationsAsRead();

    const handleNotificationClick = useCallback(
        (notificationId: string, isRead: boolean) => {
            if (!isRead) markAsRead.mutate(notificationId);
            setIsOpen(false);
        },
        [markAsRead],
    );

    if (!user) {
        return (
            <Link
                href={`/${lng}/login`}
                className="hover:bg-muted relative cursor-pointer rounded-md p-1.5 transition-colors"
            >
                <IconBell className="size-5" />
            </Link>
        );
    }

    return (
        <Popover open={isOpen} onOpenChange={setIsOpen}>
            <PopoverTrigger asChild>
                <button className="hover:bg-muted relative cursor-pointer rounded-md p-1.5 transition-colors">
                    <IconBell className="size-5" />
                    {unreadCount > 0 && (
                        <span className="bg-destructive absolute -top-0.5 -right-0.5 flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[10px] font-bold text-white">
                            {unreadCount > 99 ? "99+" : unreadCount}
                        </span>
                    )}
                </button>
            </PopoverTrigger>
            <PopoverContent
                className="border-border bg-muted mr-4 w-80 p-0"
                align="end"
            >
                <NotificationPopupContent
                    notifications={notifications}
                    isLoading={isLoading}
                    unreadCount={unreadCount}
                    viewAllHref={`/${lng}/notifications`}
                    t={t}
                    onMarkAllAsRead={() => markAllAsRead.mutate()}
                    onNotificationClick={handleNotificationClick}
                    onClose={() => setIsOpen(false)}
                />
            </PopoverContent>
        </Popover>
    );
}
