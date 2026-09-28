"use client";

import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { cn } from "@/utils/cn";

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
    return (
        <Avatar className={cn(userAvatarVariants({ size }), className)}>
            <AvatarImage
                src={src ?? undefined}
                alt={alt ?? name ?? undefined}
                className="object-cover"
            />
            <AvatarFallback className={fallbackClassName}>
                {fallback ?? (name || "U").charAt(0).toUpperCase()}
            </AvatarFallback>
            {children}
        </Avatar>
    );
}
