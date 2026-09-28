import { MediaCardGridSkeleton } from "@/components/skeletons/media-card-skeleton";
import { Skeleton } from "@/components/ui/skeleton";

// Shown in the content panel only; the header and sidebar stay in place.
export default function MainContentLoading() {
    return (
        <div className="flex w-full flex-col gap-4 px-8 py-4">
            <Skeleton className="my-2 h-6 w-40" />
            <MediaCardGridSkeleton count={8} />
        </div>
    );
}
