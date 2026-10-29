"use client";

import { PublicUser, SocialMediaLinks } from "@/types/user";
import FollowerCount from "@/components/user/follower-count";
import IconCalendar from "@/components/icons/calendar";
import IconUsers from "@/components/icons/users";
import IconFacebook from "@/components/icons/facebook";
import IconTwitter from "@/components/icons/twitter";
import IconInstagram from "@/components/icons/instagram";
import IconLinkedin from "@/components/icons/linkedin";
import IconGithub from "@/components/icons/github";
import IconYoutube from "@/components/icons/youtube";
import IconGlobe from "@/components/icons/globe";
import IconTiktok from "@/components/icons/tiktok";
import useT from "@/hooks/use-translation";
import { formatLocaleDate } from "@/utils/timeFormats";

const platformIcons: Record<keyof SocialMediaLinks, typeof IconGlobe> = {
    facebook: IconFacebook,
    twitter: IconTwitter,
    instagram: IconInstagram,
    linkedin: IconLinkedin,
    github: IconGithub,
    youtube: IconYoutube,
    website: IconGlobe,
    tiktok: IconTiktok,
};

type SocialLink = {
    platform: keyof SocialMediaLinks;
    href: string;
    label: string;
};

// Links are user-supplied: only http(s) URLs are rendered.
function toSocialLinks(links: SocialMediaLinks | undefined): SocialLink[] {
    return Object.entries(links ?? {}).flatMap(([platform, url]) => {
        if (!url || !(platform in platformIcons)) return [];
        try {
            const parsed = new URL(url);
            if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
                return [];
            }
            const path = parsed.pathname === "/" ? "" : parsed.pathname;
            return [
                {
                    platform: platform as keyof SocialMediaLinks,
                    href: parsed.href,
                    label: `${parsed.hostname.replace(/^www\./, "")}${path}`,
                },
            ];
        } catch {
            return [];
        }
    });
}

export default function ProfileAboutTab({ user }: { user: PublicUser }) {
    const { t, i18n } = useT("users");
    const links = toSocialLinks(user.socialMediaLinks);

    return (
        <div className="grid gap-8 @3xl:grid-cols-[minmax(0,1fr)_18rem]">
            <section aria-labelledby="channel-about">
                <h2 id="channel-about" className="mb-2 text-lg font-semibold">
                    {t("users:profile.about")} {user.username}
                </h2>
                {user.bio ? (
                    <p className="text-sm leading-relaxed whitespace-pre-line">
                        {user.bio}
                    </p>
                ) : (
                    <p className="text-muted-foreground text-sm">
                        {t("users:channel.no_bio", {
                            username: user.username,
                        })}
                    </p>
                )}
            </section>

            <aside className="bg-muted/40 border-border h-fit space-y-5 rounded-lg border p-4">
                <div>
                    <h3 className="text-muted-foreground mb-2 text-xs font-semibold tracking-wide uppercase">
                        {t("users:channel.stats")}
                    </h3>
                    <ul className="space-y-2 text-sm">
                        <li className="flex items-center gap-2">
                            <IconUsers className="text-muted-foreground" />
                            <FollowerCount count={user.followerCount} />
                        </li>
                        <li className="flex items-center gap-2">
                            <IconCalendar className="text-muted-foreground" />
                            <span>
                                {t("users:profile.joined_prefix")}{" "}
                                {formatLocaleDate(
                                    new Date(user.createdAt),
                                    i18n.resolvedLanguage,
                                )}
                            </span>
                        </li>
                    </ul>
                </div>

                {links.length > 0 && (
                    <div>
                        <h3 className="text-muted-foreground mb-2 text-xs font-semibold tracking-wide uppercase">
                            {t("users:channel.links")}
                        </h3>
                        <ul className="space-y-2 text-sm">
                            {links.map(({ platform, href, label }) => {
                                const Icon = platformIcons[platform];
                                return (
                                    <li key={platform}>
                                        <a
                                            href={href}
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            className="hover:text-primary flex items-center gap-2 hover:underline"
                                        >
                                            <Icon className="text-muted-foreground shrink-0" />
                                            <span className="truncate">
                                                {label}
                                            </span>
                                        </a>
                                    </li>
                                );
                            })}
                        </ul>
                    </div>
                )}
            </aside>
        </div>
    );
}
