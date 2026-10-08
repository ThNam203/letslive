"use client";

import Link from "next/link";
import { PublicUser } from "@/types/user";
import useT from "@/hooks/use-translation";
import UserAvatar from "@/components/ui/user-avatar";
import ProfileActions from "@/components/user/profile-actions";

/** The uploader's avatar, name and follower count with follow and gift actions. */
export default function VODChannelRow({
    user,
    updateUser,
}: {
    user: PublicUser;
    updateUser: (newUserInfo: PublicUser) => void;
}) {
    const { t, i18n } = useT(["common", "accessibility"]);
    const profileHref = `/users/${user.id}`;

    const compact = new Intl.NumberFormat(i18n.resolvedLanguage, {
        notation: "compact",
    });

    return (
        <div className="flex min-w-0 items-center gap-3">
            <Link href={profileHref} className="shrink-0">
                <UserAvatar
                    src={user.profilePicture}
                    name={user.username}
                    alt={t("accessibility:user_avatar")}
                />
            </Link>
            <div className="flex min-w-0 flex-col">
                <Link
                    href={profileHref}
                    className="text-foreground truncate font-semibold hover:underline"
                >
                    {user.username}
                </Link>
                <span className="text-muted-foreground text-xs">
                    {t("common:vod.followers", {
                        count: user.followerCount,
                        formatted: compact.format(user.followerCount),
                    })}
                </span>
            </div>
            <div className="ml-2 shrink-0">
                <ProfileActions user={user} updateUser={updateUser} />
            </div>
        </div>
    );
}
