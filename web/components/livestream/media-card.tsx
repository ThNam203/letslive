"use client";

import { useEffect, useState } from "react";
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
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import IconEye from "../icons/eye";
import IconEyeOff from "../icons/eye-off";
import IconClock from "../icons/clock";
import IconDotsVertical from "../icons/dots-vertical";
import { Button } from "../ui/button";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import { GetUserById } from "@/lib/api/user";
import { toast } from "@/components/utils/toast";

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
            ? ["common", "settings", "accessibility", "api-response", "fetch-error"]
            : ["common", "api-response", "fetch-error"],
    );

    const userId = isLive ? props.livestream.userId : props.vod.userId;
    const id = isLive ? props.livestream.id : props.vod.id;
    const title = isLive ? props.livestream.title : props.vod.title;
    const viewCount = isLive
        ? props.livestream.viewCount
        : props.vod.viewCount;
    const thumbnailUrl =
        (isLive ? props.livestream.thumbnailUrl : props.vod.thumbnailUrl) ??
        `${GLOBAL.API_URL}/files/livestreams/${id}/thumbnail.jpeg`;

    const providedUser = props.user;
    const [user, setUser] = useState<PublicUser | null>(providedUser ?? null);

    useEffect(() => {
        if (providedUser !== undefined) {
            setUser(providedUser ?? null);
            return;
        }
        const needsUser = isLive || isWithUser;
        if (!needsUser) return;
        let cancelled = false;
        GetUserById(userId)
            .then((res) => {
                if (cancelled) return;
                if (res.success) {
                    setUser(res.data ?? null);
                } else {
                    toast(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                        type: "error",
                    });
                }
            })
            .catch(() => {
                if (cancelled) return;
                toast(t("fetch-error:client_fetch_error"), {
                    toastId: "client-fetch-error-id",
                    type: "error",
                });
            });
        return () => {
            cancelled = true;
        };
    }, [userId, isLive, isWithUser, providedUser, t]);

    const goToTarget = () => {
        if (isLive) {
            router.push(`/users/${userId}`);
        } else {
            router.push(`/users/${userId}/vods/${id}`);
        }
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
                    fallbackSrc="/images/streaming.jpg"
                    className="cursor-pointer"
                    onClick={goToTarget}
                />
            );
        }
        const vod = props.vod;
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
                <LiveImage
                    src={thumbnailUrl}
                    alt={title}
                    className={cn(
                        "h-full w-full",
                        !isEditable && "hover:cursor-pointer",
                    )}
                    width={500}
                    height={500}
                    onClick={isEditable ? undefined : goToTarget}
                    fallbackSrc="/images/streaming.jpg"
                    alwaysRefresh={false}
                />
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
                {viewCount} {viewCount < 2 ? "view" : "views"}
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
        <div className="text-muted-foreground mt-1 flex items-center gap-3 text-xs">
            <div className="flex items-center gap-1">
                <IconEye className="h-3 w-3" />
                {renderViewCount()}
            </div>
            <div className="flex items-center gap-1">
                <IconClock className="h-3 w-3" />
                {renderTimeMeta()}
            </div>
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
                        <Avatar>
                            <AvatarImage
                                src={user?.profilePicture}
                                alt={`${user?.username} avatar`}
                                className="h-full w-full object-cover"
                                width={40}
                                height={40}
                            />
                            <AvatarFallback>
                                {(user?.username ?? "U")
                                    .charAt(0)
                                    .toUpperCase()}
                            </AvatarFallback>
                        </Avatar>
                    </button>
                    <div className="min-w-0 flex-1">
                        <h3
                            className="line-clamp-2 text-base font-semibold hover:cursor-pointer"
                            onClick={goToTarget}
                        >
                            {title}
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
                <div className="flex items-center gap-2">
                    <h3 className="text-foreground line-clamp-2 flex-1 text-base font-semibold">
                        {title}
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
                "border-border w-full overflow-hidden rounded-sm transition-all hover:shadow-md",
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
