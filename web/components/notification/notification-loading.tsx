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
                "text-muted-foreground flex items-center justify-center py-8 text-sm",
                className,
            )}
        >
            {message}
        </div>
    );
}
