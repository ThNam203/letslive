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

                <span className="text-muted-foreground shrink-0 pt-0.5 text-xs">
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
