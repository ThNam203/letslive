"use client";

import { MediaCardGridSkeleton } from "@/components/skeletons/media-card-skeleton";
import { Button } from "@/components/ui/button";
import IconFilm from "@/components/icons/film";
import IconRefresh from "@/components/icons/refresh";
import useT from "@/hooks/use-translation";

export const CHANNEL_GRID_CLASS =
    "grid grid-cols-1 gap-x-4 gap-y-6 @lg:grid-cols-2 @2xl:grid-cols-3 @5xl:grid-cols-4";

export function ChannelVodsSkeleton({ count = 6 }: { count?: number }) {
    return (
        <MediaCardGridSkeleton
            count={count}
            withUser={false}
            className={CHANNEL_GRID_CLASS}
        />
    );
}

export function ChannelVodsEmpty({ username }: { username: string }) {
    const { t } = useT("users");
    return (
        <div className="border-border flex flex-col items-center rounded-lg border border-dashed px-6 py-14 text-center">
            <div className="bg-muted mb-4 rounded-full p-4">
                <IconFilm className="text-muted-foreground h-8 w-8" />
            </div>
            <h3 className="text-lg font-semibold">
                {t("users:channel.no_videos")}
            </h3>
            <p className="text-muted-foreground mt-1 max-w-sm text-sm">
                {t("users:channel.no_videos_description", { username })}
            </p>
        </div>
    );
}

export function ChannelVodsError({ onRetry }: { onRetry: () => void }) {
    const { t } = useT("users");
    return (
        <div
            role="alert"
            className="border-border flex flex-col items-center gap-3 rounded-lg border px-6 py-10 text-center"
        >
            <p className="text-sm">{t("users:channel.videos_error")}</p>
            <Button variant="outline" size="sm" onClick={onRetry}>
                <IconRefresh />
                {t("users:channel.retry")}
            </Button>
        </div>
    );
}
