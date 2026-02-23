"use client";

import { cn } from "@/utils/cn";

type NotificationLoadingProps = {
    message: string;
    className?: string;
};

export function NotificationLoading({
    message,
    className,
}: NotificationLoadingProps) {
    return (
        <div
            className={cn(
                "flex items-center justify-center py-8 text-sm text-muted-foreground",
                className,
            )}
        >
            {message}
        </div>
    );
}
