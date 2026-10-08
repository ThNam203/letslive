"use client";

import Link from "next/link";
import { PublicUser } from "@/types/user";
import useT from "@/hooks/use-translation";
import UserAvatar from "@/components/ui/user-avatar";
import FollowerCount from "@/components/user/follower-count";
import ProfileActions from "@/components/user/profile-actions";

/** The uploader's avatar, name and follower count with follow and gift actions. */
export default function VODChannelRow({
    user,
    updateUser,
}: {
    user: PublicUser;
    updateUser: (newUserInfo: PublicUser) => void;
}) {
    const { t } = useT("accessibility");
    const profileHref = `/users/${user.id}`;

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
                <FollowerCount
                    count={user.followerCount}
                    className="text-muted-foreground text-xs"
                />
            </div>
            <div className="ml-2 shrink-0">
                <ProfileActions user={user} updateUser={updateUser} />
            </div>
        </div>
    );
}
