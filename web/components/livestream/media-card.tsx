"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { dateDiffFromNow, formatSeconds } from "@/utils/timeFormats";
import GLOBAL from "../../global";
import { VOD } from "@/types/vod";
import { Livestream } from "@/types/livestream";
import { PublicUser } from "@/types/user";
import LiveImage from "./live-image";
import { Hover3DBox } from "./hover-3d-box";
import { cn } from "@/utils/cn";
import useT from "@/hooks/use-translation";
import { Card, CardContent } from "../ui/card";
import { Badge } from "../ui/badge";
import UserAvatar from "@/components/ui/user-avatar";
import IconEyeOff from "../icons/eye-off";
import IconDotsVertical from "../icons/dots-vertical";
import { Button } from "../ui/button";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import { usePublicUser } from "@/hooks/queries/use-users";
import useCompactNumber from "@/hooks/use-compact-number";

// Matches the 1/2/3/4-column card grids (sm/md/lg breakpoints).
const THUMBNAIL_SIZES =
    "(min-width: 1024px) 25vw, (min-width: 768px) 33vw, (min-width: 640px) 50vw, 100vw";

export type VODVariant = "default" | "with-user" | "editable";

export type MediaCardProps =
    | {
          kind: "live";
          livestream: Livestream;
          user?: PublicUser | null;
          className?: string;
      }
    | {
          kind: "vod";
          vod: VOD;
          variant?: VODVariant;
          className?: string;
          onEdit?: () => void;
          onDelete?: () => void;
          user?: PublicUser | null;
          preload?: boolean;
      };

export default function MediaCard(props: MediaCardProps) {
    const router = useRouter();
    const isLive = props.kind === "live";
    const variant: VODVariant = isLive
        ? "with-user"
        : (props.variant ?? "default");
    const isEditable = variant === "editable";
    const isWithUser = variant === "with-user";

    const { t } = useT(
        isEditable
            ? [
                  "common",
                  "settings",
                  "accessibility",
                  "api-response",
                  "fetch-error",
              ]
            : ["common", "api-response", "fetch-error"],
    );

    const userId = isLive ? props.livestream.userId : props.vod.userId;
    const id = isLive ? props.livestream.id : props.vod.id;
    const title = isLive ? props.livestream.title : props.vod.title;
    const viewCount = isLive ? props.livestream.viewCount : props.vod.viewCount;
    const thumbnailUrl =
        (isLive ? props.livestream.thumbnailUrl : props.vod.thumbnailUrl) ??
        `${GLOBAL.API_URL}/files/livestreams/${id}/thumbnail.jpeg`;

    // A card either receives the user from its parent (lists that already
    // fetched them) or looks it up itself; only live cards and the
    // with-user variant show one at all.
    const providedUser = props.user;
    const needsLookup = providedUser === undefined && (isLive || isWithUser);
    const { data: fetchedUser } = usePublicUser(userId, needsLookup);
    const user = providedUser ?? (needsLookup ? (fetchedUser ?? null) : null);

    const formatCompact = useCompactNumber();

    const targetHref = isLive
        ? `/users/${userId}`
        : `/users/${userId}/vods/${id}`;

    const goToTarget = () => {
        router.push(targetHref);
    };

    const goToUser = (e: React.MouseEvent) => {
        e.stopPropagation();
        router.push(`/users/${userId}`);
    };

    const renderThumbnail = () => {
        if (isLive) {
            return (
                <Hover3DBox
                    showStream={true}
                    imageSrc={thumbnailUrl}
                    fallbackSrc="/images/streaming.webp"
                    className="cursor-pointer"
                    onClick={goToTarget}
                />
            );
        }
        const vod = props.vod;
        const image = (
            <LiveImage
                src={thumbnailUrl}
                alt={title}
                className={cn(
                    "h-full w-full object-cover",
                    !isEditable &&
                        "transition-transform duration-300 group-hover:scale-105",
                )}
                width={500}
                height={500}
                sizes={THUMBNAIL_SIZES}
                preload={props.preload ?? false}
                fallbackSrc="/images/streaming.webp"
                alwaysRefresh={false}
            />
        );
        return (
            <div className="bg-muted relative aspect-video overflow-hidden">
                <div className="absolute right-2 bottom-2 z-10 flex items-center gap-2">
                    {isEditable && vod.visibility !== "public" && (
                        <Badge
                            variant="secondary"
                            className="bg-destructive flex h-6 items-center justify-center px-2.5 text-white"
                        >
                            <IconEyeOff className="h-4 w-4" />
                        </Badge>
                    )}
                    <Badge
                        variant="secondary"
                        className="h-6 bg-black/70 text-white"
                    >
                        {formatSeconds(vod.duration)}
                    </Badge>
                </div>
                {isEditable ? (
                    image
                ) : (
                    <Link
                        href={targetHref}
                        tabIndex={-1}
                        aria-hidden="true"
                        className="block h-full w-full"
                    >
                        {image}
                    </Link>
                )}
            </div>
        );
    };

    const renderViewCount = () => {
        if (isEditable && !isLive) {
            return (
                <span>
                    {viewCount}{" "}
                    {t(
                        `settings:vods.metadata.${viewCount === 1 ? "view" : "views"}`,
                    )}
                </span>
            );
        }
        return (
            <span>
                {t("common:vod.views_compact", {
                    count: viewCount,
                    formatted: formatCompact(viewCount),
                })}
            </span>
        );
    };

    const renderTimeMeta = () => {
        if (isLive) {
            return (
                <span>
                    {t("common:started_at", {
                        time: dateDiffFromNow(props.livestream.startedAt, t),
                    })}
                </span>
            );
        }
        return <span>{dateDiffFromNow(props.vod.createdAt, t)}</span>;
    };

    const renderMetaRow = () => (
        <div className="text-muted-foreground mt-1 flex items-center gap-1.5 truncate text-xs">
            {renderViewCount()}
            <span aria-hidden="true">·</span>
            {renderTimeMeta()}
        </div>
    );

    const renderEditMenu = () => {
        if (!isEditable || isLive) return null;
        const { onEdit, onDelete } = props;
        if (!onEdit && !onDelete) return null;
        return (
            <DropdownMenu>
                <DropdownMenuTrigger asChild>
                    <Button
                        asChild
                        className="hover:cursor-pointer"
                        variant="ghost"
                        size="icon"
                        aria-label={t("accessibility:open_menu")}
                    >
                        <IconDotsVertical className="h-6 w-6 p-1" />
                    </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                    align="end"
                    className="border-border bg-background"
                >
                    {onEdit && (
                        <DropdownMenuItem onClick={onEdit}>
                            {t("settings:vods.actions.edit")}
                        </DropdownMenuItem>
                    )}
                    {onDelete && (
                        <DropdownMenuItem onClick={onDelete}>
                            {t("settings:vods.actions.delete")}
                        </DropdownMenuItem>
                    )}
                </DropdownMenuContent>
            </DropdownMenu>
        );
    };

    const renderContent = () => {
        if (isWithUser) {
            return (
                <div className="flex items-start gap-3">
                    <button
                        type="button"
                        onClick={goToUser}
                        aria-label={user?.username ?? "user"}
                        className="bg-muted h-10 w-10 flex-shrink-0 overflow-hidden rounded-full hover:cursor-pointer"
                    >
                        <UserAvatar
                            src={user?.profilePicture}
                            name={user?.username}
                            alt={`${user?.username} avatar`}
                        />
                    </button>
                    <div className="min-w-0 flex-1">
                        <h3 className="line-clamp-2 text-base font-semibold">
                            <Link
                                href={targetHref}
                                className="group-hover:text-primary focus-visible:underline focus-visible:outline-none"
                            >
                                {title}
                            </Link>
                        </h3>
                        <p
                            className="text-muted-foreground truncate text-sm hover:cursor-pointer hover:underline"
                            onClick={goToUser}
                        >
                            {user?.username ?? "Unknown"}
                        </p>
                        {renderMetaRow()}
                    </div>
                </div>
            );
        }
        return (
            <div>
                <div className="flex items-start gap-2">
                    <h3
                        className="text-foreground line-clamp-2 flex-1 text-base leading-snug font-semibold"
                        title={title}
                    >
                        {isEditable ? (
                            title
                        ) : (
                            <Link
                                href={targetHref}
                                className="group-hover:text-primary focus-visible:underline focus-visible:outline-none"
                            >
                                {title}
                            </Link>
                        )}
                    </h3>
                    {renderEditMenu()}
                </div>
                {renderMetaRow()}
            </div>
        );
    };

    return (
        <Card
            className={cn(
                "group border-border w-full overflow-hidden rounded-sm transition-all hover:shadow-md",
                props.className,
            )}
        >
            {renderThumbnail()}
            <CardContent
                className={cn(
                    "p-4 pt-2",
                    isWithUser && !isLive && "h-28",
                    isLive && "bg-muted",
                )}
            >
                {renderContent()}
            </CardContent>
        </Card>
    );
}
