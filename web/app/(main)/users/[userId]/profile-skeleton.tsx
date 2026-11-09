import { Skeleton } from "@/components/ui/skeleton";
import { ChannelVodsSkeleton } from "./channel-vod-states";

export function ProfileTabsSkeleton() {
    return (
        <div aria-busy="true">
            <div className="border-border flex gap-6 border-b pb-3">
                <Skeleton className="h-5 w-16" />
                <Skeleton className="h-5 w-14" />
                <Skeleton className="h-5 w-20" />
            </div>
            <div className="@container mt-6">
                <ChannelVodsSkeleton count={3} />
            </div>
        </div>
    );
}

export default function ProfileSkeleton({ className }: { className?: string }) {
    return (
        <div className={className} aria-busy="true">
            <div className="flex items-center gap-4 py-4">
                <Skeleton className="h-16 w-16 shrink-0 rounded-full" />
                <div className="flex-1 space-y-2">
                    <Skeleton className="h-7 w-48" />
                    <Skeleton className="h-4 w-64" />
                </div>
                <Skeleton className="h-9 w-24 rounded-full" />
            </div>
            <ProfileTabsSkeleton />
        </div>
    );
}
