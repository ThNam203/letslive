"use client";

import Link from "next/link";
import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { dateDiffFromNow, formatSeconds } from "@/utils/timeFormats";
import GLOBAL from "@/global";
import { VOD } from "@/types/vod";
import { User } from "@/types/user";
import LiveImage from "@/components/livestream/live-image";
import { cn } from "@/utils/cn";
import useT from "@/hooks/use-translation";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import IconEye from "@/components/icons/eye";
import IconEyeOff from "@/components/icons/eye-off";
import IconClock from "@/components/icons/clock";
import IconDotsVertical from "@/components/icons/dots-vertical";
import { Button } from "@/components/ui/button";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { GetUserById } from "@/lib/api/user";

export type VODCardVariant = "default" | "with-user" | "editable";

export interface VODCardProps {
    vod: VOD;
    variant?: VODCardVariant;
    className?: string;
    onEdit?: () => void;
    onDelete?: () => void;
    user?: User | null;
}

export default function VODCard({
    vod,
    variant = "default",
    className,
    onEdit,
    onDelete,
    user: providedUser,
}: VODCardProps) {
    const { t } = useT(
        variant === "editable"
            ? ["common", "settings", "accessibility"]
            : "common",
    );
    const router = useRouter();
    const [user, setUser] = useState<User | null>(providedUser ?? null);

    useEffect(() => {
        if (variant === "with-user" && !providedUser) {
            const fetchUser = async () => {
                const res = await GetUserById(vod.userId);
                if (res.success) setUser(res.data ?? null);
            };
            fetchUser();
        }
    }, [vod.userId, variant, providedUser]);

    const thumbnailUrl =
        vod.thumbnailUrl ??
        `${GLOBAL.API_URL}/files/livestreams/${vod.id}/thumbnail.jpeg`;

    const handleRedirectToVODClick = () => {
        router.push(`/users/${vod.userId}/vods/${vod.id}`);
    };

    const isEditable = variant === "editable";
    const isWithUser = variant === "with-user";

    const renderThumbnail = () => (
        <div className="bg-muted relative aspect-video overflow-hidden">
            <div className="absolute bottom-2 right-2 z-10 flex items-center gap-2">
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
                alt={vod.title}
                className={cn(
                    "h-full w-full",
                    !isEditable && "hover:cursor-pointer",
                )}
                width={500}
                height={500}
                onClick={handleRedirectToVODClick}
                fallbackSrc="/images/streaming.jpg"
                alwaysRefresh={false}
            />
        </div>
    );

    const renderViewCount = () => {
        if (isEditable) {
            return (
                <span>
                    {vod.viewCount}{" "}
                    {t(
                        `settings:vods.metadata.${vod.viewCount === 1 ? "view" : "views"}`,
                    )}
                </span>
            );
        }
        return (
            <span>
                {vod.viewCount} {vod.viewCount < 2 ? "view" : "views"}
            </span>
        );
    };

    const renderContent = () => (
        <div>
            {isWithUser ? (
                <div className="flex items-start gap-3">
                    <div className="bg-muted h-10 w-10 flex-shrink-0 overflow-hidden rounded-full">
                        <Avatar>
                            <AvatarImage
                                src={user?.profilePicture}
                                alt={`${user?.username} avatar`}
                                className="h-full w-full object-cover"
                                width={40}
                                height={40}
                            />
                            <AvatarFallback>
                                {user?.username.charAt(0).toUpperCase()}
                            </AvatarFallback>
                        </Avatar>
                    </div>
                    <div className="min-w-0 flex-1">
                        <h3
                            className="line-clamp-2 text-base font-semibold hover:cursor-pointer"
                            onClick={handleRedirectToVODClick}
                        >
                            {vod.title}
                        </h3>
                        <p className="text-muted-foreground truncate text-sm">
                            {user
                                ? (user.displayName ?? user.username)
                                : "Unknown"}
                        </p>
                        <div className="text-muted-foreground mt-1 flex items-center gap-3 text-xs">
                            <div className="flex items-center gap-1">
                                <IconEye className="h-3 w-3" />
                                {renderViewCount()}
                            </div>
                            <div className="flex items-center gap-1">
                                <IconClock className="h-3 w-3" />
                                <span>{dateDiffFromNow(vod.createdAt, t)}</span>
                            </div>
                        </div>
                    </div>
                </div>
            ) : (
                <>
                    <div className="flex items-center gap-2">
                        <h3 className="text-foreground line-clamp-2 flex-1 text-base font-semibold">
                            {vod.title}
                        </h3>
                        {isEditable && (onEdit || onDelete) && (
                            <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <Button
                                        asChild
                                        className="hover:cursor-pointer"
                                        variant="ghost"
                                        size="icon"
                                        aria-label={t(
                                            "accessibility:open_menu",
                                        )}
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
                        )}
                    </div>
                    <div className="text-muted-foreground mt-1 flex items-center gap-3 text-xs">
                        <div className="flex items-center gap-1">
                            <IconEye className="h-3 w-3" />
                            {renderViewCount()}
                        </div>
                        <div className="flex items-center gap-1">
                            <IconClock className="h-3 w-3" />
                            <span>{dateDiffFromNow(vod.createdAt, t)}</span>
                        </div>
                    </div>
                </>
            )}
        </div>
    );

    return (
        <Card
            className={cn(
                "rounded-xs border-border w-full overflow-hidden transition-all hover:shadow-md",
                className,
            )}
        >
            {renderThumbnail()}
            <CardContent className="h-28 p-4 pt-2">
                {renderContent()}
            </CardContent>
        </Card>
    );
}
