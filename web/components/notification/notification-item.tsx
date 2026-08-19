"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { cn } from "@/utils/cn";
import { Notification } from "@/types/notification";
import { timeAgo, type TimeAgoTranslator } from "./utils";
import useT from "@/hooks/use-translation";
import { formatLocaleDate } from "@/utils/timeFormats";

const TIME_AGO_REFRESH_MS = 60 * 1000;

type NotificationItemProps = {
    notification: Notification;
    t: TimeAgoTranslator;
    onClick?: () => void;
    /** Compact row for popup (no type badge, single line actions) */
    variant?: "compact" | "full";
};

export function NotificationItemContent({
    notification,
    t,
    variant = "compact",
}: Omit<NotificationItemProps, "onClick">) {
    const { i18n } = useT("notification");
    const [, setTick] = useState(0);

    useEffect(() => {
        const id = setInterval(() => {
            setTick((n) => n + 1);
        }, TIME_AGO_REFRESH_MS);
        return () => clearInterval(id);
    }, []);

    const fullTimestamp = formatLocaleDate(
        new Date(notification.createdAt),
        i18n.resolvedLanguage,
        {
            year: "numeric",
            month: "short",
            day: "numeric",
            hour: "2-digit",
            minute: "2-digit",
        },
    );

    return (
        <div
            className={cn(
                "flex flex-col gap-1 transition-colors",
                variant === "compact" &&
                    "border-border hover:bg-background cursor-pointer border-b px-4 py-3",
            )}
        >
            <div className="flex items-start justify-between gap-3">
                <div className="flex min-w-0 items-start gap-2">
                    <span className="flex h-6 w-2 shrink-0 items-center justify-center">
                        {!notification.isRead && (
                            <span className="bg-primary h-2 w-2 rounded-full" />
                        )}
                    </span>

                    <div className="text-foreground text-sm leading-relaxed font-medium">
                        {variant === "full" && (
                            <span className="bg-muted text-muted-foreground mr-2 inline-flex items-center rounded-md px-1.5 py-0.5 text-[10px] tracking-wide">
                                {notification.type}
                            </span>
                        )}
                        <span>{notification.title}</span>
                    </div>
                </div>

                <span
                    className="text-muted-foreground shrink-0 pt-0.5 text-xs"
                    title={fullTimestamp}
                >
                    {timeAgo(notification.createdAt, t)}
                </span>
            </div>

            <p className="text-muted-foreground pl-4 text-sm">
                {notification.message}
            </p>

            {notification.actionLabel && variant === "compact" && (
                <span className="text-primary pl-4 text-xs font-medium">
                    {notification.actionLabel}
                </span>
            )}
        </div>
    );
}

export function NotificationItem({
    notification,
    t,
    onClick,
    variant = "compact",
}: NotificationItemProps) {
    const content = (
        <NotificationItemContent
            notification={notification}
            t={t}
            variant={variant}
        />
    );

    const wrapperClassName = cn(
        variant === "compact" && "hover:bg-background",
        !notification.isRead && "bg-primary/5",
    );

    if (variant === "compact" && notification.actionUrl) {
        return (
            <Link
                href={notification.actionUrl}
                onClick={onClick}
                className={wrapperClassName}
            >
                {content}
            </Link>
        );
    }

    if (variant === "compact" && onClick) {
        return (
            <div onClick={onClick} className={wrapperClassName}>
                {content}
            </div>
        );
    }

    return <div className={wrapperClassName}>{content}</div>;
}
