import { MediaCardGridSkeleton } from "@/components/skeletons/media-card-skeleton";
import { Skeleton } from "@/components/ui/skeleton";

export default function ProfileSkeleton({
    showRecentActivity = true,
    className,
}: {
    showRecentActivity?: boolean;
    className?: string;
}) {
    return (
        <div className={className} aria-busy="true">
            <Skeleton className="h-[300px] w-full rounded-sm" />
            <div className="-mt-16 px-4 sm:-mt-24">
                <Skeleton className="border-background h-20 w-20 rounded-full border-4 sm:h-32 sm:w-32" />
            </div>

            <div className="mt-4 flex w-full flex-col gap-4 px-4 pb-8">
                <Skeleton className="h-8 w-56" />
                <div className="flex flex-row gap-2">
                    <div className="flex max-w-72 flex-1 flex-col gap-2">
                        <Skeleton className="h-3 w-32" />
                        <Skeleton className="h-3 w-40" />
                        <Skeleton className="h-3 w-24" />
                    </div>
                    <div className="flex-1 space-y-2">
                        <Skeleton className="h-5 w-24" />
                        <Skeleton className="h-4 w-full" />
                        <Skeleton className="h-4 w-2/3" />
                    </div>
                </div>
                {showRecentActivity && (
                    <div>
                        <Skeleton className="mb-4 h-6 w-40" />
                        <MediaCardGridSkeleton withUser={false} />
                    </div>
                )}
            </div>
        </div>
    );
}
