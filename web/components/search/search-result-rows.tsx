"use client";

import Link from "next/link";
import { PublicUser } from "@/types/user";
import { Livestream } from "@/types/livestream";
import { VOD } from "@/types/vod";
import UserAvatar from "@/components/ui/user-avatar";
import LiveImage from "@/components/livestream/live-image";
import FollowerCount from "@/components/user/follower-count";
import useT from "@/hooks/use-translation";
import useCompactNumber from "@/hooks/use-compact-number";
import { usePublicUser } from "@/hooks/queries/use-users";
import { dateDiffFromNow, formatSeconds } from "@/utils/timeFormats";
import { excerptAroundMatch } from "@/utils/highlight";
import { mediaThumbnailUrl } from "@/utils/thumbnail";
import { cn } from "@/utils/cn";
import { Skeleton } from "@/components/ui/skeleton";
import { SearchItem } from "@/hooks/queries/use-search";
import { SearchType } from "@/lib/search";
import Highlight from "./highlight";

type RowSize = "compact" | "full";

export type RowProps<T> = {
    item: T;
    query: string;
    size?: RowSize;
    onNavigate?: () => void;
};

const ROW_CLASS =
    "hover:bg-muted focus-visible:ring-ring flex w-full items-start gap-3 rounded-md focus-visible:ring-2 focus-visible:outline-none";

export function UserResultRow({
    item,
    query,
    size = "full",
    onNavigate,
}: RowProps<PublicUser>) {
    const compact = size === "compact";
    return (
        <Link
            href={`/users/${item.id}`}
            onClick={onNavigate}
            className={cn(ROW_CLASS, "items-center", compact ? "p-2" : "p-3")}
        >
            <UserAvatar
                size={compact ? "sm" : "lg"}
                src={item.profilePicture}
                name={item.username}
                alt=""
                className={compact ? undefined : "h-14 w-14 sm:h-14 sm:w-14"}
            />
            <div className="min-w-0 flex-1">
                <Highlight
                    text={item.username}
                    query={query}
                    className={cn(
                        "block truncate font-semibold",
                        compact ? "text-sm" : "text-base",
                    )}
                />
                <FollowerCount
                    count={item.followerCount}
                    className="text-muted-foreground block text-xs"
                />
                {!compact && item.bio && (
                    <p className="text-muted-foreground mt-0.5 truncate text-sm">
                        {item.bio}
                    </p>
                )}
            </div>
        </Link>
    );
}

function MediaRow({
    href,
    query,
    size,
    onNavigate,
    thumbnail,
    title,
    description,
    userId,
    meta,
}: {
    href: string;
    query: string;
    size: RowSize;
    onNavigate?: () => void;
    thumbnail: React.ReactNode;
    title: string;
    description: string | null;
    userId: string;
    meta: React.ReactNode;
}) {
    const { data: streamer } = usePublicUser(userId);
    const compact = size === "compact";
    const snippet =
        !compact && description ? excerptAroundMatch(description, query) : null;

    return (
        <Link
            href={href}
            onClick={onNavigate}
            className={cn(ROW_CLASS, compact ? "p-2" : "gap-4 p-3")}
        >
            <div
                className={cn(
                    "bg-muted relative aspect-video shrink-0 overflow-hidden rounded-md",
                    compact ? "w-24" : "w-36 sm:w-64",
                )}
            >
                {thumbnail}
            </div>
            <div className="min-w-0 flex-1">
                <Highlight
                    text={title}
                    query={query}
                    className={cn(
                        "font-semibold",
                        compact
                            ? "line-clamp-1 text-sm"
                            : "line-clamp-2 text-base leading-snug",
                    )}
                />
                <div
                    className={cn(
                        "text-muted-foreground flex min-w-0 items-center gap-1.5 text-xs",
                        compact ? "min-h-4" : "mt-1 min-h-5",
                    )}
                >
                    {!compact && (
                        <UserAvatar
                            size="sm"
                            src={streamer?.profilePicture}
                            name={streamer?.username}
                            alt=""
                            className="h-5 w-5"
                        />
                    )}
                    {streamer && (
                        <Highlight
                            text={streamer.username}
                            query={query}
                            className="truncate"
                        />
                    )}
                </div>
                <div className="text-muted-foreground mt-0.5 truncate text-xs">
                    {meta}
                </div>
                {snippet && (
                    <Highlight
                        text={snippet}
                        query={query}
                        className="text-muted-foreground mt-1.5 line-clamp-2 text-sm"
                    />
                )}
            </div>
        </Link>
    );
}

function Thumbnail({ id, url }: { id: string; url: string | null }) {
    return (
        <LiveImage
            src={mediaThumbnailUrl(id, url)}
            alt=""
            className="h-full w-full object-cover"
            width={320}
            height={180}
            sizes="(min-width: 640px) 256px, 160px"
            fallbackSrc="/images/streaming.webp"
            alwaysRefresh={false}
        />
    );
}

export function LiveResultRow({
    item,
    query,
    size = "full",
    onNavigate,
}: RowProps<Livestream>) {
    const { t } = useT(["common", "users"]);
    const formatCompact = useCompactNumber();
    return (
        <MediaRow
            href={`/users/${item.userId}`}
            query={query}
            size={size}
            onNavigate={onNavigate}
            title={item.title}
            description={item.description}
            userId={item.userId}
            thumbnail={
                <>
                    <Thumbnail id={item.id} url={item.thumbnailUrl} />
                    <span className="absolute top-1 left-1 rounded bg-red-600 px-1 text-[10px] leading-4 font-bold text-white uppercase">
                        {t("users:live")}
                    </span>
                </>
            }
            meta={
                <>
                    {t("common:search.viewers", {
                        count: item.viewCount,
                        formatted: formatCompact(item.viewCount),
                    })}
                    {size === "full" && (
                        <>
                            <span aria-hidden="true"> · </span>
                            {t("common:started_at", {
                                time: dateDiffFromNow(item.startedAt, t),
                            })}
                        </>
                    )}
                </>
            }
        />
    );
}

export function VodResultRow({
    item,
    query,
    size = "full",
    onNavigate,
}: RowProps<VOD>) {
    const { t } = useT("common");
    const formatCompact = useCompactNumber();
    return (
        <MediaRow
            href={`/users/${item.userId}/vods/${item.id}`}
            query={query}
            size={size}
            onNavigate={onNavigate}
            title={item.title}
            description={item.description}
            userId={item.userId}
            thumbnail={
                <>
                    <Thumbnail id={item.id} url={item.thumbnailUrl} />
                    <span className="absolute right-1 bottom-1 rounded bg-black/75 px-1 text-[10px] leading-4 font-semibold text-white">
                        {formatSeconds(item.duration)}
                    </span>
                </>
            }
            meta={
                <>
                    {t("common:vod.views_compact", {
                        count: item.viewCount,
                        formatted: formatCompact(item.viewCount),
                    })}
                    <span aria-hidden="true"> · </span>
                    {dateDiffFromNow(item.createdAt, t)}
                </>
            }
        />
    );
}

export type SearchRow<T extends SearchType> = (
    props: RowProps<SearchItem[T]>,
) => React.ReactElement;

export const SEARCH_ROWS: { [T in SearchType]: SearchRow<T> } = {
    users: UserResultRow,
    lives: LiveResultRow,
    vods: VodResultRow,
};

export function SearchRowSkeleton() {
    return (
        <div className="flex items-start gap-4 p-3">
            <Skeleton className="aspect-video w-36 shrink-0 rounded-md sm:w-64" />
            <div className="flex-1 space-y-2 pt-1">
                <Skeleton className="h-4 w-3/4" />
                <Skeleton className="h-3 w-1/3" />
                <Skeleton className="h-3 w-1/2" />
            </div>
        </div>
    );
}
