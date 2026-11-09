"use client";

import { PublicUser } from "@/types/user";
import MediaCard from "@/components/livestream/media-card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import IconChevronRight from "@/components/icons/chevron-right";
import useT from "@/hooks/use-translation";
import { usePublicVodsOfUserInfinite } from "@/hooks/queries/use-vods";
import { flattenPages } from "@/lib/query/paginated";
import { cn } from "@/utils/cn";
import FeaturedVod from "./featured-vod";
import {
    CHANNEL_GRID_CLASS,
    ChannelVodsEmpty,
    ChannelVodsError,
    ChannelVodsSkeleton,
} from "./channel-vod-states";

const MORE_VIDEOS_COUNT = 4;

export default function ProfileHomeTab({
    user,
    onViewAll,
}: {
    user: PublicUser;
    onViewAll: () => void;
}) {
    const { t } = useT("users");
    const { data, isLoading, isError, refetch } = usePublicVodsOfUserInfinite(
        user.id,
    );
    const vods = flattenPages(data);

    if (isLoading) {
        return (
            <div className="space-y-8" aria-busy="true">
                <div className="grid gap-4 @2xl:grid-cols-[minmax(0,1.5fr)_minmax(0,1fr)] @2xl:gap-6">
                    <Skeleton className="aspect-video w-full rounded-lg" />
                    <div className="flex flex-col justify-center gap-3">
                        <Skeleton className="h-6 w-4/5" />
                        <Skeleton className="h-4 w-1/3" />
                        <Skeleton className="h-4 w-full" />
                        <Skeleton className="h-9 w-24" />
                    </div>
                </div>
                <ChannelVodsSkeleton count={3} />
            </div>
        );
    }

    if (isError) {
        return <ChannelVodsError onRetry={() => refetch()} />;
    }

    const [latest, ...rest] = vods;
    if (!latest) {
        return <ChannelVodsEmpty username={user.username} />;
    }
    const more = rest.slice(0, MORE_VIDEOS_COUNT);

    return (
        <div className="space-y-10">
            <section aria-labelledby="channel-latest-video">
                <h2
                    id="channel-latest-video"
                    className="mb-3 text-lg font-semibold"
                >
                    {t("users:channel.latest_video")}
                </h2>
                <FeaturedVod vod={latest} />
            </section>

            {more.length > 0 && (
                <section aria-labelledby="channel-more-videos">
                    <div className="mb-3 flex items-center justify-between">
                        <h2
                            id="channel-more-videos"
                            className="text-lg font-semibold"
                        >
                            {t("users:channel.more_videos")}
                        </h2>
                        <Button variant="ghost" size="sm" onClick={onViewAll}>
                            {t("users:channel.view_all")}
                            <IconChevronRight />
                        </Button>
                    </div>
                    <div className={CHANNEL_GRID_CLASS}>
                        {more.map((vod, idx) => (
                            <MediaCard
                                key={vod.id}
                                kind="vod"
                                vod={vod}
                                className={cn(
                                    idx === MORE_VIDEOS_COUNT - 1 &&
                                        "hidden @5xl:block",
                                )}
                            />
                        ))}
                    </div>
                </section>
            )}
        </div>
    );
}
