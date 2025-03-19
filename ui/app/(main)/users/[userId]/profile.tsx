"use client";

import { CalendarDays, Users, Heart, ShieldCheck } from "lucide-react";
import { User } from "../../../../types/user";
import ProfileHeader from "./profile_header";
import VODLink from "../../../../components/vodlink";
import { Livestream } from "../../../../types/livestream";
import { useEffect, useState } from "react";
import { GetAllLivestreamOfUser, GetAllVODsAsAuthor } from "../../../../lib/api/livestream";
import { toast } from "react-toastify";

export default function ProfileView({
    user,
    vods,
    updateUser,
    showRecentActivity = true
}: {
    user: User;
    vods: Livestream[];
    updateUser: (newUserInfo: User) => void;
    showRecentActivity?: boolean;
}) {
    return (
        <div>
            <ProfileHeader user={user} updateUser={updateUser} />

            {/* Profile Content */}
            <div className="w-full px-4 sm:px-6 lg:px-8 pt-32 pb-16">
                <div className="flex items-start gap-8">
                    <div>
                        <h1 className="text-3xl font-bold text-gray-900">
                            {user.displayName ?? user.username}
                        </h1>
                        <p className="text-gray-500">
                            @{user.username}
                            <span className="inline-block align-middle ml-1">
                                {user?.isVerified && (
                                    <ShieldCheck
                                        color="#10b981"
                                        className="p-[1px]"
                                    />
                                )}
                            </span>
                        </p>
                    </div>
                </div>
                {/* Bio */}
                <div className="mt-2">
                    <h2 className="text-xl font-semibold text-gray-900">
                        About
                    </h2>
                    <p className="text-gray-700">{user.bio}</p>
                </div>

                {/* User Stats */}
                <div className="mt-2 flex space-x-6">
                    <div className="flex items-center text-gray-500">
                        <Users className="w-5 h-5 mr-2" />
                        <span>
                            {user.followerCount !== undefined
                                ? `${user.followerCount} follower${user.followerCount > 1 ? "s" : ""
                                }`
                                : "0 follower"}
                        </span>
                    </div>
                    <div className="flex items-center text-gray-500">
                        <CalendarDays className="w-5 h-5 mr-2" />
                        <span>
                            Joined {new Date(user.createdAt).toLocaleString()}
                        </span>
                    </div>
                </div>

                {/* Recent Activity */}
                {showRecentActivity ?
                    vods.length > 0 && (
                        <div className="mt-4">
                            <h2 className="text-xl font-semibold text-gray-900 mb-4">
                                Recent Streams
                            </h2>

                            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                                {vods.map((vod, idx) => {
                                    if (vod.status == "live") return null;
                                    return <VODLink key={idx} vod={vod} />;
                                })}
                            </div>
                        </div>
                    )
                    : null}
            </div>
        </div>
    );
}
