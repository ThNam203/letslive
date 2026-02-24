"use client";

import { PublicUser } from "@/types/user";
import ProfileHeader from "./profile-header";
import VODCard from "@/components/livestream/vod-card";
import IconCalendar from "@/components/icons/calendar";
import IconUsers from "@/components/icons/users";
import { VOD } from "@/types/vod";
import useT from "@/hooks/use-translation";
import IconFacebook from "@/components/icons/facebook";
import IconTwitter from "@/components/icons/twitter";
import IconInstagram from "@/components/icons/instagram";
import IconLinkedin from "@/components/icons/linkedin";
import IconGithub from "@/components/icons/github";
import IconYoutube from "@/components/icons/youtube";
import IconGlobe from "@/components/icons/globe";
import IconTiktok from "@/components/icons/tiktok";
import Link from "next/link";

const platformOptions = {
    facebook: IconFacebook,
    twitter: IconTwitter,
    instagram: IconInstagram,
    linkedin: IconLinkedin,
    github: IconGithub,
    youtube: IconYoutube,
    website: IconGlobe,
    tiktok: IconTiktok,
};

export default function ProfileView({
    user,
    vods,
    updateUser,
    showRecentActivity = true,
    className,
}: {
    user: PublicUser;
    vods: VOD[];
    updateUser: (newUserInfo: PublicUser) => void;
    showRecentActivity?: boolean;
    className?: string;
}) {
    const { t } = useT("users");

    return (
        <div className={className}>
            <ProfileHeader user={user} updateUser={updateUser} />
            {/* Profile Content */}
            <div className="mt-4 flex w-full flex-col gap-4 px-4 pb-8">
                <div className="flex items-start gap-8">
                    <div>
                        <h1 className="text-foreground text-3xl font-bold">
                            {user.displayName ?? user.username}
                        </h1>
                        <p className="text-foreground-muted">
                            @{user.username}
                        </p>
                    </div>
                </div>
                <div className="flex flex-row gap-2">
                    <div className="text-foreground flex max-w-72 flex-1 flex-col gap-2 text-xs">
                        <div className="flex items-center gap-2">
                            <IconUsers />
                            <p>
                                <span className="font-bold">
                                    {user.followerCount}
                                </span>{" "}
                                {t(
                                    user.followerCount === 1
                                        ? "users:profile.followers_one"
                                        : "users:profile.followers_other",
                                )}
                            </p>
                        </div>
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
                                        className="flex items-center gap-2"
                                        aria-label={platform}
                                    >
                                        <Icon />
                                        <span>{url}</span>
                                    </Link>
                                );
                            },
                        )}
                        <div className="flex items-center gap-2">
                            <IconCalendar />
                            <span>
                                {t("users:profile.joined_prefix")}{" "}
                                {new Date(user.createdAt).toLocaleString()}
                            </span>
                        </div>
                    </div>

                    <div className="flex-1">
                        <h2 className="text-lg font-semibold">
                            {t("users:profile.about")}
                        </h2>
                        <p className="text-sm">{user.bio}</p>
                    </div>
                </div>

                {/* Recent Activity */}
                {showRecentActivity
                    ? vods.length > 0 && (
                          <div>
                              <h2 className="text-foreground mb-4 text-xl font-semibold">
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
