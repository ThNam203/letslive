"use client";

import Image from "next/image";
import { PublicUser } from "@/types/user";
import { Livestream } from "@/types/livestream";
import UserAvatar from "@/components/ui/user-avatar";
import FollowerCount from "@/components/user/follower-count";
import ProfileActions from "@/components/user/profile-actions";
import useT from "@/hooks/use-translation";
import { formatLocaleDate } from "@/utils/timeFormats";
import { cn } from "@/utils/cn";

// Live channels already show the player above, so the banner is offline-only.
export default function ProfileHeader({
    user,
    livestream,
    updateUser,
}: {
    user: PublicUser;
    livestream: Livestream | null;
    updateUser: (newUserInfo: PublicUser) => void;
}) {
    const { t, i18n } = useT(["users", "accessibility"]);
    const isLive = livestream !== null;

    return (
        <header className="@container">
            {!isLive && <OfflineBanner user={user} />}
            <div className="flex flex-col gap-3 py-4 @xl:flex-row @xl:items-center @xl:gap-4">
                <div className="flex min-w-0 flex-1 items-center gap-4">
                    <div className="relative shrink-0">
                        <UserAvatar
                            size="lg"
                            src={user.profilePicture}
                            name={user.username}
                            alt={t("accessibility:user_avatar")}
                            className={cn(
                                "h-16 w-16 sm:h-16 sm:w-16",
                                isLive &&
                                    "ring-offset-background ring-2 ring-red-600 ring-offset-2",
                            )}
                        />
                        {isLive && (
                            <span className="absolute -bottom-1.5 left-1/2 -translate-x-1/2 rounded bg-red-600 px-1.5 py-px text-[10px] leading-tight font-bold whitespace-nowrap text-white uppercase">
                                {t("users:live")}
                            </span>
                        )}
                    </div>

                    <div className="min-w-0 flex-1">
                        <h1 className="text-foreground truncate text-xl font-bold @md:text-2xl">
                            {user.username}
                        </h1>
                        {isLive && livestream.title && (
                            <p className="truncate text-sm font-medium">
                                {livestream.title}
                            </p>
                        )}
                        <p className="text-muted-foreground flex flex-wrap gap-x-1.5 text-sm">
                            <FollowerCount
                                count={user.followerCount}
                                className="whitespace-nowrap"
                            />
                            <span aria-hidden="true">·</span>
                            <span className="whitespace-nowrap">
                                {t("users:profile.joined_prefix")}{" "}
                                {formatLocaleDate(
                                    new Date(user.createdAt),
                                    i18n.resolvedLanguage,
                                )}
                            </span>
                        </p>
                    </div>
                </div>

                <ProfileActions user={user} updateUser={updateUser} />
            </div>
        </header>
    );
}

function OfflineBanner({ user }: { user: PublicUser }) {
    const { t } = useT(["users", "accessibility"]);
    return (
        <div className="from-primary/30 via-muted to-muted relative mt-1 h-28 w-full overflow-hidden rounded-lg bg-gradient-to-br @xl:h-40">
            {user.backgroundPicture && (
                <Image
                    src={user.backgroundPicture}
                    alt={t("accessibility:profile_banner")}
                    className="object-cover"
                    fill={true}
                    sizes="(min-width: 1024px) 60vw, 100vw"
                    preload={true}
                />
            )}
            <span className="absolute bottom-3 left-3 flex items-center gap-1.5 rounded-full bg-black/65 px-3 py-1 text-xs font-semibold text-white">
                <span className="h-2 w-2 rounded-full bg-gray-400" />
                {t("users:offline")}
            </span>
        </div>
    );
}
