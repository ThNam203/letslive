"use client";

import { User } from "@/src/types/user";
import ProfileHeader from "./profile-header";
import VODCard from "@/src/components/livestream/vod-card";
import IconCalendar from "@/src/components/icons/calendar";
import IconUsers from "@/src/components/icons/users";
import { VOD } from "@/src/types/vod";
import useT from "@/src/hooks/use-translation";
import IconFacebook from "@/src/components/icons/facebook";
import IconTwitter from "@/src/components/icons/twitter";
import IconInstagram from "@/src/components/icons/instagram";
import IconLinkedin from "@/src/components/icons/linkedin";
import IconGithub from "@/src/components/icons/github";
import IconYoutube from "@/src/components/icons/youtube";
import IconGlobe from "@/src/components/icons/globe";
import Link from "next/link";
import { useTheme } from "next-themes";

const platformOptions = {
    facebook: IconFacebook,
    twitter: IconTwitter,
    instagram: IconInstagram,
    linkedin: IconLinkedin,
    github: IconGithub,
    youtube: IconYoutube,
    website: IconGlobe,
};

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
    const { t } = useT("users");
    const { resolvedTheme } = useTheme();

    const getIconTheme = (label: string) => {
        const mainColor = resolvedTheme === "light" ? "white" : "transparent";
        const color = resolvedTheme === "light" ? "black" : "white";

        return label === "Facebook"
            ? {
                  mainColor,
                  color,
              }
            : { color };
    };

    return (
        <div className={className}>
            <ProfileHeader user={user} updateUser={updateUser} />
            {/* Profile Content */}
            <div className="mt-4 w-full px-4 pb-16">
                <div className="flex items-start gap-8">
                    <div>
                        <h1 className="text-3xl font-bold text-foreground">
                            {user.displayName ?? user.username}
                        </h1>
                        <p className="text-foreground-muted">
                            @{user.username}
                        </p>
                    </div>
                </div>
                {/* Bio */}
                <div className="mt-2">
                    <h2 className="text-xl font-semibold text-foreground">
                        {t("users:profile.about")}
                    </h2>
                    <p className="text-foreground-muted">{user.bio}</p>
                </div>

                {/* User Stats */}
                <div className="mt-2 flex space-x-6">
                    <div className="flex items-center text-foreground-muted">
                        <IconUsers className="mr-2" />
                        <span>
                            {user.followerCount !== undefined
                                ? `${user.followerCount} ${t(user.followerCount === 1 ? "users:profile.followers_one" : "users:profile.followers_other")}`
                                : `0 ${t("users:profile.followers_other")}`}
                        </span>
                    </div>
                    <div className="flex items-center text-foreground-muted">
                        <IconCalendar className="mr-2" />
                        <span>
                            {t("users:profile.joined_prefix")}{" "}
                            {new Date(user.createdAt).toLocaleString()}
                        </span>
                    </div>
                </div>

                <nav className="mt-4 flex space-x-4">
                    {Object.entries(user.socialMediaLinks ?? {}).map(
                        ([platform, url]) => {
                            const Icon =
                                platformOptions[
                                    platform as keyof typeof platformOptions
                                ];
                            if (!Icon || !url) return null;
                            return (
                                <Link
                                    key={platform}
                                    href={url}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="text-foreground"
                                    aria-label={platform}
                                >
                                    <Icon
                                        className="h-5 w-5"
                                        {...getIconTheme(platform)}
                                    />
                                </Link>
                            );
                        },
                    )}
                </nav>

                {/* Recent Activity */}
                {showRecentActivity
                    ? vods.length > 0 && (
                          <div className="mt-4">
                              <h2 className="mb-4 text-xl font-semibold text-foreground">
                                  {t("users:profile.recent_streams")}
                              </h2>

                              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
                                  {vods.map((vod, idx) => {
                                      return <VODCard key={idx} vod={vod} />;
                                  })}
                              </div>
                          </div>
                      )
                    : null}
            </div>
        </div>
    );
}
