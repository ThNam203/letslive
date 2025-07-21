"use client";

import { User } from "../../../../types/user";
import ProfileHeader from "./profile_header";
import VODView from "../../../../components/livestream/vod";
import IconCalendar from "@/components/icons/calendar";
import IconUsers from "@/components/icons/users";
import { Button } from "@/components/ui/button";
import { useState } from "react";
import { toast } from "react-toastify";
import useUser from "../../../../hooks/user";
import { FollowOtherUser, UnfollowOtherUser } from "../../../../lib/api/user";
import IconLoader from "@/components/icons/loader";
import { VOD } from "@/types/vod";

export default function ProfileView({
  user,
  vods,
  updateUser,
  showRecentActivity = true,
  className,
}: {
  user: User;
  vods: VOD[];
  updateUser: (newUserInfo: User) => void;
  showRecentActivity?: boolean;
  className?: string;
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
    <div className={className}>
      <ProfileHeader user={user} updateUser={updateUser} />
      {/* Profile Content */}
      <div className="w-full px-4 sm:px-6 lg:px-8 pt-32 pb-16">
        <div className="flex items-start gap-8">
          <div>
            <h1 className="text-3xl font-bold text-foreground">
              {user.displayName ?? user.username}
            </h1>
            <p className="text-foreground-muted">@{user.username}</p>
          </div>
        </div>
        {/* Bio */}
        <div className="mt-2">
          <h2 className="text-xl font-semibold text-foreground">About</h2>
          <p className="text-foreground-muted">{user.bio}</p>
        </div>

        {me?.id !== user.id && (
          <Button
            variant={user.isFollowing ? "destructive" : "default"}
            disabled={isFetching || !me}
            onClick={onFollow}
          >
            {isFetching && <IconLoader className="mr-1" />}
            {user.isFollowing ? "Unfollow" : "Follow"}
          </Button>
        )}

        {/* User Stats */}
        <div className="mt-2 flex space-x-6">
          <div className="flex items-center text-foreground-muted">
            <IconUsers className="mr-2" />
            <span>
              {user.followerCount !== undefined
                ? `${user.followerCount} follower${
                    user.followerCount > 1 ? "s" : ""
                  }`
                : "0 follower"}
            </span>
          </div>
          <div className="flex items-center text-foreground-muted">
            <IconCalendar className="mr-2" />
            <span>Joined {new Date(user.createdAt).toLocaleString()}</span>
          </div>
        </div>

        {/* Recent Activity */}
        {showRecentActivity
          ? vods.length > 0 && (
              <div className="mt-4">
                <h2 className="text-xl font-semibold text-foreground mb-4">
                  Recent Streams
                </h2>

                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                  {vods.map((vod, idx) => {
                    return <VODView key={idx} vod={vod} />;
                  })}
                </div>
              </div>
            )
          : null}
      </div>
    </div>
  );
}
