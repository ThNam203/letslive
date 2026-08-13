"use client";

import Image from "next/image";
import { PublicUser } from "@/types/user";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import useUser from "@/hooks/user";
import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { FollowOtherUser, UnfollowOtherUser } from "@/lib/api/user";
import { toast } from "@/components/utils/toast";
import { Button } from "@/components/ui/button";
import IconLoader from "@/components/icons/loader";
import useT from "@/hooks/use-translation";
import GiftModal from "./gift-modal";

export default function ProfileHeader({
    user,
    updateUser,
}: {
    user: PublicUser;
    updateUser: (newUserInfo: PublicUser) => void;
}) {
    const { t } = useT([
        "common",
        "users",
        "fetch-error",
        "api-response",
        "accessibility",
        "shop",
    ]);
    const me = useUser((state) => state.user);
    const [isGiftModalOpen, setIsGiftModalOpen] = useState(false);

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

    const onFollowClick = () => {
        followMutation.mutate();
    };

    return (
        <div className="relative">
            <div className="relative h-[300px] w-full overflow-hidden rounded-sm bg-gray-100 shadow">
                {/* Profile Banner */}
                <Image
                    src={
                        user.backgroundPicture ??
                        `https://placehold.co/1200x600/F3F4F6/374151/png?font=playfair-display&text=${
                            user.username || "User"
                        }`
                    }
                    alt={t("accessibility:profile_banner")}
                    className="object-cover"
                    fill={true}
                    priority={true}
                    unoptimized
                />
            </div>
            <div className="-mt-16 px-4 sm:-mt-24">
                <div className="relative inline-block">
                    <Avatar className="h-32 w-32 rounded-full border-4 border-white">
                        <AvatarImage
                            src={user.profilePicture}
                            alt={t("accessibility:user_avatar")}
                        />
                        <AvatarFallback>
                            {(user.username || "U")[0].toUpperCase()}
                        </AvatarFallback>
                    </Avatar>
                    {me?.id && me.id !== user.id && (
                        <>
                            <Button
                                variant={user.isFollowing ? "destructive" : "default"}
                                disabled={followMutation.isPending || !me}
                                onClick={onFollowClick}
                                className="absolute right-0 bottom-4 flex translate-x-[50%] flex-row items-center justify-center gap-0"
                            >
                                {followMutation.isPending && (
                                    <IconLoader className="mr-1" />
                                )}
                                {user.isFollowing ? t("common:unfollow") : t("common:follow")}
                            </Button>
                            <Button
                                variant="outline"
                                onClick={() => setIsGiftModalOpen(true)}
                                className="absolute right-0 bottom-16 flex translate-x-[50%] flex-row items-center justify-center gap-1"
                            >
                                🎁 {t("shop:shop.gift_button")}
                            </Button>
                            <GiftModal
                                open={isGiftModalOpen}
                                onClose={() => setIsGiftModalOpen(false)}
                                recipientUserId={user.id}
                                recipientName={user.username}
                            />
                        </>
                    )}
                </div>
            </div>
        </div>
    );
}
