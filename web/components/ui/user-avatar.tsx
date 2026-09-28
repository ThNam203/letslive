"use client";

import * as React from "react";
import Image from "next/image";
import { cva, type VariantProps } from "class-variance-authority";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { cn } from "@/utils/cn";

type UserAvatarSize = "sm" | "md" | "lg";

const userAvatarVariants = cva("shrink-0", {
    variants: {
        size: {
            sm: "h-8 w-8",
            md: "h-10 w-10",
            lg: "h-20 w-20 sm:h-32 sm:w-32",
        },
    },
    defaultVariants: {
        size: "md",
    },
});

const IMAGE_SIZES: Record<UserAvatarSize, string> = {
    sm: "32px",
    md: "40px",
    lg: "(min-width: 640px) 128px, 80px",
};

type UserAvatarProps = VariantProps<typeof userAvatarVariants> & {
    src?: string | null;
    name?: string | null;
    alt?: string;
    fallback?: React.ReactNode;
    className?: string;
    fallbackClassName?: string;
    children?: React.ReactNode;
};

export default function UserAvatar({
    src,
    name,
    alt,
    size,
    fallback,
    className,
    fallbackClassName,
    children,
}: UserAvatarProps) {
    const [failedSrc, setFailedSrc] = React.useState<string | null>(null);
    const showImage = Boolean(src) && src !== failedSrc;

    return (
        <Avatar className={cn(userAvatarVariants({ size }), className)}>
            <AvatarFallback
                delayMs={0}
                className={cn("border-border border", fallbackClassName)}
            >
                {fallback ?? (name || "U").charAt(0).toUpperCase()}
            </AvatarFallback>
            {showImage && src && (
                <Image
                    src={src}
                    alt={alt ?? name ?? ""}
                    fill
                    sizes={IMAGE_SIZES[size ?? "md"]}
                    className="object-cover"
                    // blob: previews, and http://localhost storage in dev,
                    // can't go through the optimizer.
                    unoptimized={!src.startsWith("https://")}
                    onError={() => setFailedSrc(src)}
                />
            )}
            {children}
        </Avatar>
    );
}
