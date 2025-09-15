"use client";

import Image from "next/image";
import { User } from "@/types/user";
import {
    Avatar,
    AvatarFallback,
    AvatarImage,
} from "@/components/ui/avatar";
import useUser from "@/hooks/user";
import { useState } from "react";
import { FollowOtherUser, UnfollowOtherUser } from "@/lib/api/user";
import { toast } from "react-toastify";
import { Button } from "@/components/ui/button";
import IconLoader from "@/components/icons/loader";

export default function ProfileHeader({
    user,
    updateUser,
}: {
    user: User;
    updateUser: (newUserInfo: User) => void;
}) {
    const me = useUser((state) => state.user);
    const [isFetching, setIsFetching] = useState(false);

    const onFollowClick = async () => {
        setIsFetching(true);

        if (user.isFollowing) {
            const { fetchError } = await UnfollowOtherUser(user.id);
            if (fetchError) {
                toast(fetchError.message, {
                    toastId: "follow-error",
                    type: "error",
                });
            } else {
                updateUser({
                    ...user,
                    isFollowing: false,
                    followerCount: user.followerCount - 1,
                });
            }
        } else {
            const { fetchError } = await FollowOtherUser(user.id);
            if (fetchError) {
                toast(fetchError.message, {
                    toastId: "follow-error",
                    type: "error",
                });
            } else {
                updateUser({
                    ...user,
                    isFollowing: true,
                    followerCount: user.followerCount + 1,
                });
            }
        }
        setIsFetching(false);
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
                    className="mx-auto h-full w-auto object-cover"
                    width="0"
                    height="0"
                    sizes="100vw"
                    priority={true}
                />
            </div>
            <div className="-mt-16 px-4 sm:-mt-24">
                <div className="relative inline-block">
                  <Avatar className="h-32 w-32 rounded-full border-4 border-white">
                      <AvatarImage src={user.profilePicture} alt="user avatar" />
                      <AvatarFallback>
                          {user.username[0].toUpperCase()}
                      </AvatarFallback>
                  </Avatar>
                  {me?.id !== user.id && (
                      <Button
                          variant={user.isFollowing ? "destructive" : "default"}
                          disabled={isFetching || !me}
                          onClick={onFollowClick}
                          className="absolute bottom-4 right-0 translate-x-[50%] flex flex-row items-center justify-center gap-0"
                      >
                          {isFetching && <IconLoader className="mr-1" />}
                          {user.isFollowing ? "Unfollow" : "Follow"}
                      </Button>
                  )}
                </div>
            </div>
        </div>
    );
}
