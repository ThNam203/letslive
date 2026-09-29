import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/utils/cn";

export function PlayerSkeleton({ className }: { className?: string }) {
    return (
        <Skeleton
            aria-busy="true"
            className={cn("aspect-video w-full rounded-none", className)}
        />
    );
}
