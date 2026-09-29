import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/utils/cn";

type NotificationLoadingProps = {
    /** read out to screen readers in place of the placeholder rows */
    message: string;
    variant?: "full" | "compact";
    rows?: number;
    className?: string;
};

export function NotificationLoading({
    message,
    variant = "compact",
    rows = 4,
    className,
}: NotificationLoadingProps) {
    return (
        <div
            role="status"
            aria-busy="true"
            className={cn(
                "flex flex-col",
                variant === "full" && "gap-2",
                className,
            )}
        >
            <span className="sr-only">{message}</span>
            {Array.from({ length: rows }, (_, i) => (
                <div
                    key={i}
                    className={cn(
                        "flex flex-col gap-2 pl-4",
                        variant === "full"
                            ? "border-border rounded-lg border p-4 pl-8"
                            : "border-border border-b px-4 py-3 pl-8",
                    )}
                >
                    <Skeleton className="h-4 w-4/5" />
                    <Skeleton className="h-3 w-1/4" />
                </div>
            ))}
        </div>
    );
}
