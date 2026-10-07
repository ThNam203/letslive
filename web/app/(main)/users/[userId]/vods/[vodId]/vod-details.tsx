"use client";

import { useEffect, useRef, useState } from "react";
import { VOD } from "@/types/vod";
import { PublicUser } from "@/types/user";
import useT from "@/hooks/use-translation";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import TimestampedText from "@/components/vod/timestamped-text";
import IconShare from "@/components/icons/share";
import { dateDiffFromNow, formatLocaleDate } from "@/utils/timeFormats";
import { cn } from "@/utils/cn";
import VODChannelRow from "./vod-channel-row";
import VODReactionButtons from "./vod-reaction-buttons";
import ShareVODDialog from "./share-vod-dialog";

/** Title, channel row, like/dislike, share and description under the player. */
export default function VODDetails({
    vod,
    user,
    isLoadingUser,
    updateUser,
    getCurrentTime,
    className,
}: {
    vod: VOD;
    user: PublicUser | undefined;
    isLoadingUser: boolean;
    updateUser: (newUserInfo: PublicUser) => void;
    getCurrentTime: () => number;
    className?: string;
}) {
    const { t, i18n } = useT("common");
    const [shareTime, setShareTime] = useState<number | null>(null);

    return (
        <section className={cn("flex flex-col gap-3 px-1", className)}>
            <h1 className="text-foreground text-xl font-semibold break-words">
                {vod.title}
            </h1>

            <div className="flex flex-wrap items-center justify-between gap-3">
                {isLoadingUser ? (
                    <div className="flex items-center gap-3">
                        <Skeleton className="h-10 w-10 rounded-full" />
                        <div className="flex flex-col gap-1">
                            <Skeleton className="h-4 w-32" />
                            <Skeleton className="h-3 w-20" />
                        </div>
                    </div>
                ) : (
                    user && (
                        <VODChannelRow user={user} updateUser={updateUser} />
                    )
                )}

                <div className="flex items-center gap-2">
                    <VODReactionButtons vod={vod} />
                    <Button
                        variant="secondary"
                        onClick={() =>
                            setShareTime(Math.floor(getCurrentTime()))
                        }
                        className="rounded-full"
                    >
                        <IconShare />
                        {t("common:vod.share")}
                    </Button>
                </div>
            </div>

            <div className="bg-muted rounded-lg p-3 text-sm">
                <p className="text-foreground font-semibold">
                    {t("common:vod.views", { count: vod.viewCount })}
                    <span aria-hidden> · </span>
                    <time
                        dateTime={vod.createdAt}
                        title={formatLocaleDate(
                            new Date(vod.createdAt),
                            i18n.resolvedLanguage,
                            {
                                year: "numeric",
                                month: "short",
                                day: "numeric",
                            },
                        )}
                    >
                        {dateDiffFromNow(vod.createdAt, t)}
                    </time>
                </p>
                {vod.description?.trim() && (
                    <VODDescription description={vod.description} />
                )}
            </div>

            {shareTime !== null && (
                <ShareVODDialog
                    open
                    onOpenChange={(open) => !open && setShareTime(null)}
                    vodPath={`/users/${vod.userId}/vods/${vod.id}`}
                    currentTime={shareTime}
                />
            )}
        </section>
    );
}

function VODDescription({ description }: { description: string }) {
    const { t } = useT("common");
    const textRef = useRef<HTMLParagraphElement>(null);
    const [expanded, setExpanded] = useState(false);
    const [overflows, setOverflows] = useState(false);

    // measured while clamped, so the toggle only shows when text is cut off
    useEffect(() => {
        const el = textRef.current;
        if (!el || expanded) return;
        const measure = () => setOverflows(el.scrollHeight > el.clientHeight);
        measure();
        const observer = new ResizeObserver(measure);
        observer.observe(el);
        return () => observer.disconnect();
    }, [description, expanded]);

    return (
        <>
            <p
                ref={textRef}
                className={cn(
                    "text-foreground mt-1 break-words whitespace-pre-wrap",
                    !expanded && "line-clamp-3",
                )}
            >
                <TimestampedText text={description} />
            </p>
            {(overflows || expanded) && (
                <button
                    type="button"
                    onClick={() => setExpanded((prev) => !prev)}
                    className="text-foreground mt-1 font-semibold hover:underline"
                >
                    {expanded
                        ? t("common:vod.show_less")
                        : t("common:vod.show_more")}
                </button>
            )}
        </>
    );
}
