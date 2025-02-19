import Image from "next/image";
import { Button } from "@/components/ui/button";
import { CalendarDays, Users, Heart } from "lucide-react";
import Link from "next/link";
import { User } from "@/types/user";
import ProfileHeader from "@/app/(main)/users/[userId]/profile_header";
import { GetUserById } from "@/lib/api/user";

export default async function ProfileView({ user }: { user: User }) {
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
                <div className="mt-10">
                    <h2 className="text-xl font-semibold text-gray-900 mb-4">
                        Recent Streams
                    </h2>
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                        {[1, 2, 3].map((item, idx) => (
                            <div
                                key={item}
                                className="bg-gray-200 overflow-hidden shadow-sm rounded-sm"
                            >
                                <Link
                                    key={idx}
                                    className={`w-full h-[180px] inline-block hover:cursor-pointer`}
                                    href={`/users/${item}/vods/${item}`}
                                >
                                    <div className="flex flex-col items-center justify-center h-full bg-black bg-opacity-50">
                                        <Image
                                            alt="vod icon"
                                            src={"/icons/video.svg"}
                                            width={100}
                                            height={100}
                                        />
                                        <p className="text-white">
                                            Streamed on {item}
                                        </p>
                                    </div>
                                </Link>
                                <div className="p-4">
                                    <h3 className="font-semibold text-gray-900">
                                        Stream Title {item}
                                    </h3>
                                    <p className="text-sm text-gray-500 mt-1">
                                        Game Name • 2 days ago
                                    </p>
                                    <div className="flex items-center mt-2 text-sm text-gray-500">
                                        <Heart className="w-4 h-4 mr-1" />
                                        <span>1.2K views</span>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
}
