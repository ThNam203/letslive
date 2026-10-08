"use client";

import Image from "next/image";
import { PublicUser } from "@/types/user";
import UserAvatar from "@/components/ui/user-avatar";
import useT from "@/hooks/use-translation";

export default function ProfileHeader({ user }: { user: PublicUser }) {
    const { t } = useT("accessibility");
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
                    sizes="100vw"
                    preload={true}
                />
            </div>
            <div className="-mt-16 px-4 sm:-mt-24">
                <div className="relative inline-block">
                    <UserAvatar
                        size="lg"
                        src={user.profilePicture}
                        name={user.username}
                        alt={t("accessibility:user_avatar")}
                        className="border-4 border-white"
                    />
                </div>
            </div>
        </div>
    );
}
