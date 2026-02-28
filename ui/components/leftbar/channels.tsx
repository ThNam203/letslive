"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import useUser from "../../hooks/user";
import { PublicUser } from "../../types/user";
import {
    GetFollowingChannels,
    GetRecommendedChannels,
} from "../../lib/api/user";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import { cn } from "@/utils/cn";
import {
    HoverCard,
    HoverCardContent,
    HoverCardTrigger,
} from "../ui/hover-card";
import useT from "@/hooks/use-translation";
import { toast } from "@/components/utils/toast";

function ChannelUserCard({
    user,
    isMinimized,
}: {
    user: PublicUser;
    isMinimized: boolean;
}) {
    const { t } = useT(["common"]);
    return (
        <HoverCard>
            <HoverCardTrigger asChild>
                <Link
                    href={`/users/${user.id}`}
                    className={cn(
                        "hover:bg-background-hover flex flex-row items-center gap-2 rounded-full",
                        isMinimized ? "" : "w-full",
                    )}
                >
                    <Avatar>
                        <AvatarImage
                            src={user.profilePicture}
                            alt="User avatar"
                            className="h-10 w-10 rounded-full"
                        />
                        <AvatarFallback className="border-border h-10 w-10 rounded-full border">
                            {user.username.charAt(0).toUpperCase()}
                        </AvatarFallback>
                    </Avatar>
                    {!isMinimized && (
                        <span className="text-sm font-semibold">
                            {user.displayName ?? user.username}
                        </span>
                    )}
                </Link>
            </HoverCardTrigger>
            <HoverCardContent className="border-border bg-muted z-10 w-80">
                <div className="flex gap-4">
                    <Avatar>
                        <AvatarImage
                            src={user.profilePicture}
                            alt="User avatar"
                            className="h-10 w-10 rounded-full"
                        />
                        <AvatarFallback className="border-border h-10 w-10 rounded-full border">
                            {user.username.charAt(0).toUpperCase()}
                        </AvatarFallback>
                    </Avatar>
                    <div className="space-y-1">
                        <h4 className="text-sm font-semibold">
                            {user.displayName ?? user.username}
                        </h4>
                        <p className="text-muted-foreground text-xs">
                            {t("common:bio")}:{" "}
                            {user.bio ?? t("common:no_description")}
                        </p>
                        <p className="text-muted-foreground text-xs">
                            {t("common:followers_with_count", {
                                count: user.followerCount,
                            })}
                        </p>
                        <p className="text-muted-foreground text-xs">
                            {t("common:joined")}:{" "}
                            {new Date(user.createdAt).toLocaleDateString()}
                        </p>
                    </div>
                </div>
            </HoverCardContent>
        </HoverCard>
    );
}

export default function AllChannelsView({
    isMinimized = false,
    minimizeLeftBarIcon,
}: {
    isMinimized?: boolean;
    minimizeLeftBarIcon?: React.ReactNode;
}) {
    const curUser = useUser((state) => state.user);
    const [followingUsers, setFollowingUsers] = useState<PublicUser[]>([]);
    const [recommendedUsers, setRecommendedUsers] = useState<PublicUser[]>([]);
    const { t } = useT(["common", "api-response", "fetch-error"]);

    useEffect(() => {
        if (!curUser) return;
        GetFollowingChannels()
            .then((res) => {
                if (res.success) setFollowingUsers(res.data ?? []);
            })
            .catch(() => {
                setFollowingUsers([]);
            });
    }, [curUser]);

    useEffect(() => {
        GetRecommendedChannels(0)
            .then((res) => {
                if (res.success) setRecommendedUsers(res.data ?? []);
                else
                    toast(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                        type: "error",
                    });
            })
            .catch(() => {
                toast(t("fetch-error:client_fetch_error"), {
                    toastId: "client-fetch-error-id",
                    type: "error",
                });
            });
    }, [t]);

    return (
        <div
            className={cn(
                "flex w-full flex-col items-center gap-2",
                isMinimized ? "" : "px-4",
            )}
        >
            <div className="flex w-full flex-row items-center justify-between">
                {!isMinimized ? (
                    <h2 className="text-xl font-semibold">{t("channels")}</h2>
                ) : null}
                {minimizeLeftBarIcon}
            </div>

            {curUser && followingUsers.length > 0 && (
                <>
                    {!isMinimized && (
                        <h3 className="text-muted-foreground w-full text-sm font-medium">
                            {t("common:following")}
                        </h3>
                    )}
                    {followingUsers.map((user) => (
                        <ChannelUserCard
                            key={user.id}
                            user={user}
                            isMinimized={isMinimized}
                        />
                    ))}
                </>
            )}

            {recommendedUsers.length > 0 && (
                <>
                    {!isMinimized && (
                        <h3 className="text-muted-foreground w-full text-sm font-medium">
                            {t("common:recommended")}
                        </h3>
                    )}
                    {recommendedUsers.map((user) => (
                        <ChannelUserCard
                            key={user.id}
                            user={user}
                            isMinimized={isMinimized}
                        />
                    ))}
                </>
            )}
        </div>
    );
}
