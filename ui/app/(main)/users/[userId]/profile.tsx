import Image from "next/image";
import { CalendarDays, Users, Heart } from "lucide-react";
import Link from "next/link";
import { User } from "@/types/user";
import ProfileHeader from "@/app/(main)/users/[userId]/profile_header";
import VODLink from "@/components/vodlink";

export default function ProfileView({ user, showSavedStream = true }: { user: User; showSavedStream?: boolean }) { 
    return (
        <div>
            {/* Profile Header */}
            <div className="relative">
                <div className="h-48 bg-gray-100 border-1 border-gray-400 rounded-md">
                    {/* Profile Banner */}
                    <Image
                        src={
                            user.backgroundPicture ??
                            `https://placehold.co/1200x800/F3F4F6/374151/png?font=playfair-display&text=${
                                user.displayName ?? user.username
                            }`
                        }
                        alt="Profile Banner"
                        width={1200}
                        height={800}
                        className="w-full h-full object-cover rounded-md"
                    />
                </div>
                <div className="max-w-5xl sm:px-6 px-4">
                    <div className="relative -mt-16 sm:-mt-24">
                        <div className="absolute">
                            {/* Profile Picture */}
                            <Image
                                src={
                                    user.profilePicture ??
                                    "https://github.com/shadcn.png"
                                }
                                alt="Profile Picture"
                                width={128}
                                height={128}
                                className="rounded-full border-4 border-white"
                            />
                        </div>
                    </div>
                </div>
            </div>

            {/* Profile Content */}
            <div className="max-w-5xl px-4 sm:px-6 lg:px-8 pt-32 pb-16">
                <ProfileHeader user={user} />
                {/* Bio */}
                <div className="mt-6">
                    <h2 className="text-xl font-semibold text-gray-900 mb-2">
                        About
                    </h2>
                    <p className="text-gray-700">{user.bio}</p>
                </div>

                {/* User Stats */}
                <div className="mt-6 flex space-x-6">
                    <div className="flex items-center text-gray-500">
                        <Users className="w-5 h-5 mr-2" />
                        <span>10.2K followers</span>
                    </div>
                    <div className="flex items-center text-gray-500">
                        <CalendarDays className="w-5 h-5 mr-2" />
                        <span>
                            Joined {new Date(user.createdAt).toLocaleString()}
                        </span>
                    </div>
                </div>

                {/* Recent Activity */}
                {showSavedStream && <div className="mt-10">
                    <h2 className="text-xl font-semibold text-gray-900 mb-4">
                        Recent Streams
                    </h2>
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                        {user.vods?.map((item, idx) => {
                            if (item.status == "live") return;
                            return <VODLink key={idx} item={item} />;
                        })}
                    </div>
                </div>}
            </div>
        </div>
    );
}
