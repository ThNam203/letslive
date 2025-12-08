"use client";

import Image from "next/image";
import { User } from "@/types/user";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import useUser from "@/hooks/user";
import { useState } from "react";
import { FollowOtherUser, UnfollowOtherUser } from "@/lib/api/user";
import { toast } from "react-toastify";
import { Button } from "@/components/ui/button";
import IconLoader from "@/components/icons/loader";
import useT from "@/hooks/use-translation";

export default function ProfileHeader({
    user,
    updateUser,
}: {
    user: User;
    updateUser: (newUserInfo: User) => void;
}) {
    const { t } = useT(["common", "users", "fetch-error", "api-response"]);
    const me = useUser((state) => state.user);
    const [isFetching, setIsFetching] = useState(false);

    const onFollowClick = async () => {
        setIsFetching(true);

        if (user.isFollowing) {
            await UnfollowOtherUser(user.id)
                .then((res) => {
                    if (res.success) {
                        updateUser({
                            ...user,
                            isFollowing: false,
                            followerCount: user.followerCount - 1,
                        });
                    } else {
                        toast(t(`api-response:${res.key}`), {
                            toastId: res.requestId,
                            type: "error",
                        });
                    }
                })
                .catch((err) => {
                    toast(t("fetch-error:client_fetch_error"), {
                        toastId: "client-fetch-error-id",
                        type: "error",
                    });
                })
                .finally(() => {
                    setIsFetching(false);
                });
        } else {
            await FollowOtherUser(user.id)
                .then((res) => {
                    if (res.success) {
                        updateUser({
                            ...user,
                            isFollowing: true,
                            followerCount: user.followerCount + 1,
                        });
                    } else {
                        toast(t(`api-response:${res.key}`), {
                            toastId: res.key,
                            type: "error",
                        });
                    }
                })
                .catch((err) => {
                    toast(t("fetch-error:client_fetch_error"), {
                        toastId: "client-fetch-error-id",
                        type: "error",
                    });
                })
                .finally(() => {
                    setIsFetching(false);
                });
        }
    };

    return (
        <div className="relative">
            <div className="relative h-[300px] w-full overflow-hidden rounded-sm bg-gray-100 shadow">
                {/* Profile Banner */}
                <Image
                    src={
                        user.backgroundPicture ??
                        `https://placehold.co/1200x600/F3F4F6/374151/png?font=playfair-display&text=${
                            user.displayName ?? user.username
                        }`
                    }
                    alt="Profile Banner"
                    className="object-cover"
                    fill={true}
                    priority={true}
                />
            </div>
            <div className="-mt-16 px-4 sm:-mt-24">
                <div className="relative inline-block">
                    <Avatar className="h-32 w-32 rounded-full border-4 border-white">
                        <AvatarImage
                            src={user.profilePicture}
                            alt="user avatar"
                        />
                        <AvatarFallback>
                            {user.username[0].toUpperCase()}
                        </AvatarFallback>
                    </Avatar>
                    {me?.id && me.id !== user.id && (
                        <Button
                            variant={
                                user.isFollowing ? "destructive" : "default"
                            }
                            disabled={isFetching || !me}
                            onClick={onFollowClick}
                            className="absolute bottom-4 right-0 flex translate-x-[50%] flex-row items-center justify-center gap-0"
                        >
                            {isFetching && <IconLoader className="mr-1" />}
                            {user.isFollowing
                                ? t("common:unfollow")
                                : t("common:follow")}
                        </Button>
                    )}
                </div>
            </div>
        </div>
    );
}
