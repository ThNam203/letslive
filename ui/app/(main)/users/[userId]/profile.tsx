"use client";

import { CalendarDays, Users, Heart, ShieldCheck } from "lucide-react";
import { User } from "../../../../types/user";
import ProfileHeader from "./profile_header";
import VODLink from "../../../../components/vodlink";

export default function ProfileView({
    user,
    showSavedStream = true,
    updateUser,
}: {
    user: User;
    showSavedStream?: boolean;
    updateUser: (newUserInfo: User) => void;
}) {
    return (
        <div>
            <ProfileHeader user={user} updateUser={updateUser} />

            {/* Profile Content */}
            <div className="max-w-5xl px-4 sm:px-6 lg:px-8 pt-32 pb-16">
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
                                ? `${user.followerCount} follower${
                                      user.followerCount > 1 ? "s" : ""
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
                {showSavedStream && (
                    <div className="mt-4">
                        <h2 className="text-xl font-semibold text-gray-900 mb-4">
                            Recent Streams
                        </h2>
                        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                            {user.vods?.map((item, idx) => {
                                if (item.status == "live") return;
                                return <VODLink key={idx} item={item} />;
                            })}
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
