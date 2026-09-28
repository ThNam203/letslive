import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/utils/cn";

/** The image + name + price card used by the shop, gifts and inventory. */
export function ItemCardSkeleton({
    imageClassName = "h-24 w-24",
    withAction = false,
}: {
    imageClassName?: string;
    withAction?: boolean;
}) {
    return (
        <div className="border-border bg-card flex flex-col items-center gap-2 rounded-xl border p-4">
            <Skeleton className={imageClassName} />
            <Skeleton className="h-4 w-3/4" />
            <Skeleton className="h-5 w-1/2 rounded-full" />
            {withAction && <Skeleton className="h-8 w-full" />}
        </div>
    );
}

export function ItemGridSkeleton({
    count = 10,
    className,
    imageClassName,
    withAction,
}: {
    count?: number;
    className: string;
    imageClassName?: string;
    withAction?: boolean;
}) {
    return (
        <div className={cn("grid", className)} aria-busy="true">
            {Array.from({ length: count }, (_, i) => (
                <ItemCardSkeleton
                    key={i}
                    imageClassName={imageClassName}
                    withAction={withAction}
                />
            ))}
        </div>
    );
}
