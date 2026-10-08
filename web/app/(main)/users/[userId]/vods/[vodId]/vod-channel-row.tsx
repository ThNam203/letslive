"use client";

import { useState } from "react";
import Link from "next/link";
import { useMutation } from "@tanstack/react-query";
import { PublicUser } from "@/types/user";
import { FollowOtherUser, UnfollowOtherUser } from "@/lib/api/user";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import UserAvatar from "@/components/ui/user-avatar";
import { Button } from "@/components/ui/button";
import IconLoader from "@/components/icons/loader";
import { toast } from "@/components/utils/toast";
import GiftModal from "../../gift-modal";

/** The uploader's avatar, name and follower count with follow and gift actions. */
export default function VODChannelRow({
    user,
    updateUser,
}: {
    user: PublicUser;
    updateUser: (newUserInfo: PublicUser) => void;
}) {
    const { t, i18n } = useT([
        "common",
        "api-response",
        "accessibility",
        "shop",
    ]);
    const me = useUser((state) => state.user);
    const [isGiftModalOpen, setIsGiftModalOpen] = useState(false);
    const profileHref = `/users/${user.id}`;

    const followMutation = useMutation({
        mutationFn: () =>
            user.isFollowing
                ? UnfollowOtherUser(user.id)
                : FollowOtherUser(user.id),
        onSuccess: (res) => {
            if (res.success) {
                updateUser({
                    ...user,
                    isFollowing: !user.isFollowing,
                    followerCount: user.isFollowing
                        ? user.followerCount - 1
                        : user.followerCount + 1,
                });
            } else {
                toast(t(`api-response:${res.key}`), {
                    toastId: res.requestId,
                    type: "error",
                });
            }
        },
    });

    const compact = new Intl.NumberFormat(i18n.resolvedLanguage, {
        notation: "compact",
    });
    const canInteract = Boolean(me?.id) && me?.id !== user.id;

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
            {canInteract && (
                <div className="ml-2 flex shrink-0 items-center gap-2">
                    <Button
                        variant={user.isFollowing ? "outline" : "default"}
                        disabled={followMutation.isPending}
                        onClick={() => followMutation.mutate()}
                        className="rounded-full"
                    >
                        {followMutation.isPending && <IconLoader />}
                        {user.isFollowing
                            ? t("common:unfollow")
                            : t("common:follow")}
                    </Button>
                    <Button
                        variant="outline"
                        onClick={() => setIsGiftModalOpen(true)}
                        className="rounded-full"
                    >
                        🎁 {t("shop:shop.gift_button")}
                    </Button>
                    <GiftModal
                        open={isGiftModalOpen}
                        onClose={() => setIsGiftModalOpen(false)}
                        recipientUserId={user.id}
                        recipientName={user.username}
                    />
                </div>
            )}
        </div>
    );
}
