"use client";

import Link from "next/link";
import useUser from "../../hooks/user";
import { PublicUser } from "../../types/user";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import { cn } from "@/utils/cn";
import {
    HoverCard,
    HoverCardContent,
    HoverCardTrigger,
} from "../ui/hover-card";
import useT from "@/hooks/use-translation";
import { formatLocaleDate } from "@/utils/timeFormats";
import {
    useFollowingChannels,
    useRecommendedChannels,
} from "@/hooks/queries/use-channels";

function ChannelUserCard({
    user,
    isMinimized,
}: {
    user: PublicUser;
    isMinimized: boolean;
}) {
    const { t, i18n } = useT(["common", "accessibility"]);
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
                            alt={t("accessibility:user_avatar")}
                            className="h-10 w-10 rounded-full"
                        />
                        <AvatarFallback className="border-border h-10 w-10 rounded-full border">
                            {(user.username ?? "U").charAt(0).toUpperCase()}
                        </AvatarFallback>
                    </Avatar>
                    {!isMinimized && (
                        <span className="text-sm font-semibold">
                            {user.username}
                        </span>
                    )}
                </Link>
            </HoverCardTrigger>
            <HoverCardContent className="border-border bg-muted z-10 w-80">
                <div className="flex gap-4">
                    <Avatar>
                        <AvatarImage
                            src={user.profilePicture}
                            alt={t("accessibility:user_avatar")}
                            className="h-10 w-10 rounded-full"
                        />
                        <AvatarFallback className="border-border h-10 w-10 rounded-full border">
                            {(user.username ?? "U").charAt(0).toUpperCase()}
                        </AvatarFallback>
                    </Avatar>
                    <div className="space-y-1">
                        <h4 className="text-sm font-semibold">
                            {user.username}
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
                            {formatLocaleDate(
                                new Date(user.createdAt),
                                i18n.resolvedLanguage,
                            )}
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
    const { data: followingUsers = [] } = useFollowingChannels(!!curUser);
    const { data: recommendedUsers = [] } = useRecommendedChannels(0);
    const { t } = useT(["common"]);

    const followingIds = new Set(followingUsers.map((user) => user.id));
    const recommendedNotFollowing = recommendedUsers.filter(
        (user) => !followingIds.has(user.id),
    );

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

            {recommendedNotFollowing.length > 0 && (
                <>
                    {!isMinimized && (
                        <h3 className="text-muted-foreground w-full text-sm font-medium">
                            {t("common:recommended")}
                        </h3>
                    )}
                    {recommendedNotFollowing.map((user) => (
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
