"use client";

import { cn } from "@/utils/cn";
import IconMessage from "@/components/icons/message";

type CommentEmptyProps = {
    message: string;
    className?: string;
};

export function CommentEmpty({ message, className }: CommentEmptyProps) {
    return (
        <div
            className={cn(
                "flex flex-col items-center justify-center gap-3 py-10 text-center",
                className,
            )}
        >
            <div
                className="flex h-12 w-12 items-center justify-center rounded-full bg-muted"
                aria-hidden
            >
                <IconMessage className="text-muted-foreground" />
            </div>
            <p className="text-sm text-muted-foreground max-w-[280px]">
                {message}
            </p>
        </div>
    );
}
