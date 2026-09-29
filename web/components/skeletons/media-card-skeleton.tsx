import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/utils/cn";

export const MEDIA_CARD_GRID_CLASS =
    "grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4";

export function MediaCardSkeleton({ withUser = true }: { withUser?: boolean }) {
    return (
        <Card className="border-border w-full overflow-hidden rounded-sm">
            <Skeleton className="aspect-video w-full rounded-none" />
            <CardContent className={cn("p-4 pt-2", withUser && "h-28")}>
                {withUser ? (
                    <div className="flex items-start gap-3 pt-2">
                        <Skeleton className="h-10 w-10 shrink-0 rounded-full" />
                        <div className="flex-1 space-y-2">
                            <Skeleton className="h-4 w-full" />
                            <Skeleton className="h-3 w-1/2" />
                            <Skeleton className="h-3 w-1/3" />
                        </div>
                    </div>
                ) : (
                    <div className="space-y-2 pt-2">
                        <Skeleton className="h-4 w-3/4" />
                        <Skeleton className="h-3 w-1/3" />
                    </div>
                )}
            </CardContent>
        </Card>
    );
}

export function MediaCardGridSkeleton({
    count = 4,
    withUser = true,
    className = MEDIA_CARD_GRID_CLASS,
}: {
    count?: number;
    withUser?: boolean;
    className?: string;
}) {
    return (
        <div className={className} aria-busy="true">
            {Array.from({ length: count }, (_, i) => (
                <MediaCardSkeleton key={i} withUser={withUser} />
            ))}
        </div>
    );
}
