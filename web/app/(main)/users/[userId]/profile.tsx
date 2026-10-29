"use client";

import { Suspense } from "react";
import { PublicUser } from "@/types/user";
import { Livestream } from "@/types/livestream";
import ProfileHeader from "./profile-header";
import ProfileTabs from "./profile-tabs";
import { ProfileTabsSkeleton } from "./profile-skeleton";

export default function ProfileView({
    user,
    livestream,
    updateUser,
    className,
}: {
    user: PublicUser;
    livestream: Livestream | null;
    updateUser: (newUserInfo: PublicUser) => void;
    className?: string;
}) {
    return (
        <div className={className}>
            <ProfileHeader
                user={user}
                livestream={livestream}
                updateUser={updateUser}
            />
            <Suspense fallback={<ProfileTabsSkeleton />}>
                <ProfileTabs user={user} />
            </Suspense>
        </div>
    );
}
