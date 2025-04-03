"use client";

import Image from "next/image";
import { Loader } from "lucide-react";
import { useState } from "react";
import { toast } from "react-toastify";
import { User } from "../../../../types/user";
import useUser from "../../../../hooks/user";
import { FollowOtherUser, UnfollowOtherUser } from "../../../../lib/api/user";
import { Button } from "../../../../components/ui/button";
import { cn } from "../../../../utils/cn";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "../../../../components/ui/avatar";

export default function ProfileHeader({
  user,
  updateUser,
}: {
  user: User;
  updateUser: (newUserInfo: User) => void;
}) {
  const me = useUser((state) => state.user);
  const [isFetching, setIsFetching] = useState(false);

  const onFollow = async () => {
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
      <div className="relative w-full h-[300px] bg-gray-100 rounded-md overflow-hidden shadow">
        {/* Profile Banner */}
        <Image
          src={
            user.backgroundPicture ??
            `https://placehold.co/1200x800/F3F4F6/374151/png?font=playfair-display&text=${
              user.displayName ?? user.username
            }`
          }
          alt="Profile Banner"
          layout="fill"
          objectFit="cover"
        />
      </div>
      <div className="sm:px-6 px-4">
        <div className="relative -mt-16 sm:-mt-24">
          <div className="absolute">
            <div className="flex flex-row items-end gap-4">
              <Avatar className="rounded-full border-4 border-white w-32 h-32">
                <AvatarImage src={user.profilePicture} alt="user avatar" />
                <AvatarFallback>
                  {user.username[0].toUpperCase()}
                </AvatarFallback>
              </Avatar>

              {me?.id !== user.id && (
                <Button
                  className={cn(
                    "text-white border-none mb-4",
                    user.isFollowing
                      ? "bg-red-400 hover:bg-red-500"
                      : "bg-purple-600 hover:bg-purple-700"
                  )}
                  disabled={isFetching}
                  onClick={onFollow}
                >
                  {isFetching && <Loader className="animate-spin mr-1" />}
                  {user.isFollowing ? "Unfollow" : "Follow"}
                </Button>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
