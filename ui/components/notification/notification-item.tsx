"use client";

import Link from "next/link";
import { cn } from "@/utils/cn";
import { Notification } from "@/types/notification";
import { timeAgo, type TimeAgoTranslator } from "./utils";

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
    return (
        <div
            className={cn(
                "flex flex-col gap-1 transition-colors",
                variant === "compact" &&
                    "cursor-pointer border-b border-border px-4 py-3 hover:bg-background",
            )}
        >
            <div className="flex items-start justify-between gap-2">
                <div className="flex items-center gap-2">
                    <span className="flex h-2 w-2 shrink-0 items-center justify-center">
                        {!notification.isRead && (
                            <span className="h-2 w-2 rounded-full bg-primary" />
                        )}
                    </span>
                    <span className="text-sm font-medium text-foreground">
                        {notification.title}
                    </span>
                    {variant === "full" && (
                        <span className="rounded-md bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">
                            {notification.type}
                        </span>
                    )}
                </div>
                <span className="shrink-0 text-xs text-muted-foreground">
                    {timeAgo(notification.createdAt, t)}
                </span>
            </div>

            <p
                className={cn(
                    "pl-4 text-muted-foreground",
                    variant === "compact" ? "text-xs" : "text-sm",
                )}
            >
                {notification.message}
            </p>

            {notification.actionLabel && variant === "compact" && (
                <span className="pl-4 text-xs font-medium text-primary">
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
