import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/utils/cn";

type AvatarSize = "sm" | "md" | "none";

const AVATAR_CLASS: Record<Exclude<AvatarSize, "none">, string> = {
    sm: "h-8 w-8",
    md: "h-10 w-10",
};

/** An avatar + text lines row, the shape of most lists in the app. */
export function ListRowSkeleton({
    avatar = "md",
    lines = 2,
    className,
}: {
    avatar?: AvatarSize;
    lines?: 0 | 1 | 2;
    className?: string;
}) {
    return (
        <div className={cn("flex items-center gap-3", className)}>
            {avatar !== "none" && (
                <Skeleton
                    className={cn(
                        "shrink-0 rounded-full",
                        AVATAR_CLASS[avatar],
                    )}
                />
            )}
            {lines > 0 && (
                <div className="flex-1 space-y-2">
                    <Skeleton className="h-4 w-2/5" />
                    {lines === 2 && <Skeleton className="h-3 w-4/5" />}
                </div>
            )}
        </div>
    );
}

export function ListSkeleton({
    rows = 5,
    className,
    rowClassName,
    avatar,
    lines,
}: {
    rows?: number;
    className?: string;
    rowClassName?: string;
    avatar?: AvatarSize;
    lines?: 0 | 1 | 2;
}) {
    return (
        <div className={cn("flex flex-col", className)} aria-busy="true">
            {Array.from({ length: rows }, (_, i) => (
                <ListRowSkeleton
                    key={i}
                    avatar={avatar}
                    lines={lines}
                    className={rowClassName}
                />
            ))}
        </div>
    );
}
