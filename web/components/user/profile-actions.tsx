"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { PublicUser } from "@/types/user";
import useUser from "@/hooks/user";
import useT from "@/hooks/use-translation";
import { FollowOtherUser, UnfollowOtherUser } from "@/lib/api/user";
import { toast } from "@/components/utils/toast";
import { Button } from "@/components/ui/button";
import IconLoader from "@/components/icons/loader";
import IconGift from "@/components/icons/gift";
import GiftModal from "./gift-modal";

export default function ProfileActions({
    user,
    updateUser,
}: {
    user: PublicUser;
    updateUser: (newUserInfo: PublicUser) => void;
}) {
    const { t } = useT(["common", "api-response", "shop"]);
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

    if (!me?.id || me.id === user.id) return null;

    return (
        <div className="flex flex-row items-center gap-2">
            <Button
                variant={user.isFollowing ? "outline" : "default"}
                disabled={followMutation.isPending}
                onClick={() => followMutation.mutate()}
                className="rounded-full"
            >
                {followMutation.isPending && <IconLoader />}
                {user.isFollowing ? t("common:unfollow") : t("common:follow")}
            </Button>
            <Button
                variant="outline"
                size="icon"
                className="rounded-full"
                onClick={() => setIsGiftModalOpen(true)}
                aria-label={t("shop:shop.gift_button")}
                title={t("shop:shop.gift_button")}
            >
                <IconGift width="1.1rem" height="1.1rem" />
            </Button>
            <GiftModal
                open={isGiftModalOpen}
                onClose={() => setIsGiftModalOpen(false)}
                recipientUserId={user.id}
                recipientName={user.username}
            />
        </div>
    );
}
