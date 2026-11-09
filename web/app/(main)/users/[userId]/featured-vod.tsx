"use client";

import Link from "next/link";
import { VOD } from "@/types/vod";
import { mediaThumbnailUrl } from "@/utils/thumbnail";
import LiveImage from "@/components/livestream/live-image";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import IconPlay from "@/components/icons/play";
import useT from "@/hooks/use-translation";
import useCompactNumber from "@/hooks/use-compact-number";
import { dateDiffFromNow, formatSeconds } from "@/utils/timeFormats";

export default function FeaturedVod({ vod }: { vod: VOD }) {
    const { t } = useT(["users", "common"]);
    const formatCompact = useCompactNumber();
    const href = `/users/${vod.userId}/vods/${vod.id}`;
    const thumbnailUrl = mediaThumbnailUrl(vod.id, vod.thumbnailUrl);

    return (
        <article className="group grid gap-4 @2xl:grid-cols-[minmax(0,1.5fr)_minmax(0,1fr)] @2xl:gap-6">
            <Link
                href={href}
                tabIndex={-1}
                aria-hidden="true"
                className="bg-muted relative block aspect-video overflow-hidden rounded-lg"
            >
                <LiveImage
                    src={thumbnailUrl}
                    alt={vod.title}
                    className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
                    width={960}
                    height={540}
                    sizes="(min-width: 1024px) 50vw, 100vw"
                    preload={true}
                    fallbackSrc="/images/streaming.webp"
                    alwaysRefresh={false}
                />
                <span className="absolute inset-0 flex items-center justify-center bg-black/0 transition-colors group-hover:bg-black/30">
                    <span className="flex h-14 w-14 scale-90 items-center justify-center rounded-full bg-black/60 text-white opacity-0 transition-all group-hover:scale-100 group-hover:opacity-100">
                        <IconPlay className="h-7 w-7" />
                    </span>
                </span>
                <Badge
                    variant="secondary"
                    className="absolute right-2 bottom-2 h-6 bg-black/70 text-white"
                >
                    {formatSeconds(vod.duration)}
                </Badge>
            </Link>

            <div className="flex min-w-0 flex-col justify-center gap-2">
                <h3 className="line-clamp-2 text-xl leading-snug font-semibold">
                    <Link
                        href={href}
                        className="group-hover:text-primary focus-visible:underline focus-visible:outline-none"
                    >
                        {vod.title}
                    </Link>
                </h3>
                <p className="text-muted-foreground text-sm">
                    {t("common:vod.views_compact", {
                        count: vod.viewCount,
                        formatted: formatCompact(vod.viewCount),
                    })}
                    <span aria-hidden="true"> · </span>
                    {dateDiffFromNow(vod.createdAt, t)}
                </p>
                {vod.description && (
                    <p className="text-muted-foreground line-clamp-3 text-sm whitespace-pre-line">
                        {vod.description}
                    </p>
                )}
                <Button asChild className="mt-2 w-fit">
                    <Link href={href}>
                        <IconPlay />
                        {t("users:channel.watch")}
                    </Link>
                </Button>
            </div>
        </article>
    );
}
