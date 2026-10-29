"use client";

import { PublicUser } from "@/types/user";
import MediaCard from "@/components/livestream/media-card";
import { Button } from "@/components/ui/button";
import IconLoader from "@/components/icons/loader";
import useT from "@/hooks/use-translation";
import { usePublicVodsOfUserInfinite } from "@/hooks/queries/use-vods";
import { flattenPages } from "@/lib/query/paginated";
import {
    CHANNEL_GRID_CLASS,
    ChannelVodsEmpty,
    ChannelVodsError,
    ChannelVodsSkeleton,
} from "./channel-vod-states";

export default function ProfileVideosTab({ user }: { user: PublicUser }) {
    const { t } = useT("users");
    const {
        data,
        isLoading,
        isError,
        refetch,
        hasNextPage,
        isFetchingNextPage,
        isFetchNextPageError,
        fetchNextPage,
    } = usePublicVodsOfUserInfinite(user.id);
    const vods = flattenPages(data);

    if (isLoading) {
        return <ChannelVodsSkeleton count={6} />;
    }

    if (isError && vods.length === 0) {
        return <ChannelVodsError onRetry={() => refetch()} />;
    }

    if (vods.length === 0) {
        return <ChannelVodsEmpty username={user.username} />;
    }

    return (
        <div>
            <div className={CHANNEL_GRID_CLASS}>
                {vods.map((vod, idx) => (
                    <MediaCard
                        key={vod.id}
                        kind="vod"
                        vod={vod}
                        preload={idx < 3}
                    />
                ))}
            </div>
            {isFetchNextPageError && (
                <div className="mt-6">
                    <ChannelVodsError onRetry={() => fetchNextPage()} />
                </div>
            )}
            {hasNextPage && !isFetchNextPageError && (
                <div className="mt-8 flex justify-center">
                    <Button
                        variant="outline"
                        onClick={() => fetchNextPage()}
                        disabled={isFetchingNextPage}
                    >
                        {isFetchingNextPage && (
                            <IconLoader className="animate-spin" />
                        )}
                        {t("users:channel.load_more")}
                    </Button>
                </div>
            )}
        </div>
    );
}
