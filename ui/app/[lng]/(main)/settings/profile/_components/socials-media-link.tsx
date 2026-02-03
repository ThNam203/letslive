"use client";

import { useState } from "react";
import { Input } from "@/components/ui/input";
import { cn } from "@/utils/cn";
import IconFacebook from "@/components/icons/facebook";
import IconTwitter from "@/components/icons/twitter";
import IconInstagram from "@/components/icons/instagram";
import IconLinkedin from "@/components/icons/linkedin";
import IconGithub from "@/components/icons/github";
import IconYoutube from "@/components/icons/youtube";
import IconGlobe from "@/components/icons/globe";
import IconTiktok from "@/components/icons/tiktok";
import IconClose from "@/components/icons/close";
import { useTheme } from "next-themes";
import useT from "@/hooks/use-translation";
import { toast } from "react-toastify";
import { UpdateProfile } from "@/lib/api/user";
import IconCheck from "@/components/icons/check";
import { SocialMediaLinks } from "@/types/user";

interface SocialMediaEditProps {
    initialLinks?: SocialMediaLinks;
}

const platformOptions = [
    {
        value: "facebook",
        label: "Facebook",
        icon: IconFacebook,
        placeholder: "https://facebook.com/username",
    },
    {
        value: "twitter",
        label: "X/Twitter",
        icon: IconTwitter,
        placeholder: "https://x.com/username",
    },
    {
        value: "instagram",
        label: "Instagram",
        icon: IconInstagram,
        placeholder: "https://instagram.com/username",
    },
    {
        value: "linkedin",
        label: "LinkedIn",
        icon: IconLinkedin,
        placeholder: "https://linkedin.com/in/username",
    },
    {
        value: "github",
        label: "GitHub",
        icon: IconGithub,
        placeholder: "https://github.com/username",
    },
    {
        value: "youtube",
        label: "YouTube",
        icon: IconYoutube,
        placeholder: "https://youtube.com/@username",
    },
    {
        value: "website",
        label: "Website",
        icon: IconGlobe,
        placeholder: "https://yourwebsite.com",
    },
    {
        value: "tiktok",
        label: "TikTok",
        icon: IconTiktok,
        placeholder: "https://tiktok.com/@username",
    },
] as const;

export function SocialMediaEdit({ initialLinks = {} }: SocialMediaEditProps) {
    const [links, setLinks] = useState<Record<string, string>>(() => {
        const linkMap: Record<string, string> = {};
        Object.entries(initialLinks).forEach(([platform, url]) => {
            if (typeof url === "string") linkMap[platform] = url;
        });
        return linkMap;
    });

    const [expandedFields, setExpandedFields] = useState<Set<string>>(
        new Set(),
    );

    const { t } = useT("common");
    const { resolvedTheme } = useTheme();

    const handleInputChange = (platform: string, value: string) => {
        setLinks((prev) => ({
            ...prev,
            [platform]: value,
        }));
    };

    const expandField = (platform: string) => {
        setExpandedFields((prev) => {
            const newSet = new Set(prev);
            newSet.add(platform);
            return newSet;
        });
    };

    const collapseField = (platform: string, clearValue = false) => {
        setExpandedFields((prev) => {
            const newSet = new Set(prev);
            newSet.delete(platform);
            return newSet;
        });

        if (clearValue) {
            setLinks((prevLinks) => {
                const newLinks = { ...prevLinks };
                delete newLinks[platform];
                return newLinks;
            });
        }
    };

    const handleSave = async (platform: string, value: string) => {
        try {
            const trimmedValue = value.trim();
            if (trimmedValue.length === 0) {
                toast.error(t("settings:social_media_links.err_empty_url"));
            }

            links[platform] = trimmedValue;

            new URL(trimmedValue); // ensures it's a valid URL
            await UpdateProfile({
                socialMediaLinks: { [platform]: trimmedValue },
            })
                .then((res) => {
                    if (res.success) {
                        toast.success(t(`api-response:${res.key}`), {
                            toastId: res.requestId,
                            type: "success",
                        });
                        collapseField(platform);
                    }
                })
                .catch((_) => {
                    toast(t("fetch-error:client_fetch_error"), {
                        toastId: "client-fetch-error-id",
                        type: "error",
                    });
                });
        } catch (e) {
            if (e instanceof TypeError) {
                toast.error(t("settings:social_media_links.invalid_url"));
            }
        }
    };

    const extractUsername = (
        url: string,
        platform: string,
    ): string | undefined => {
        try {
            const urlObj = new URL(url);
            const pathname = urlObj.pathname;
            const segments = pathname.split("/").filter(Boolean);
            if (segments.length > 0) {
                return segments[0].replace("@", "");
            }
        } catch {
            return undefined;
        }
        return undefined;
    };

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
        <div className="space-y-2">
            {platformOptions.map(
                ({ value, label, icon: Icon, placeholder }) => {
                    const isExpanded = expandedFields.has(value);
                    const hasValue = links[value] && links[value].trim() !== "";
                    const displayUsername = hasValue
                        ? extractUsername(links[value], value) || links[value]
                        : null;

                    return (
                        <div key={value} className="space-y-2">
                            {!isExpanded ? (
                                <button
                                    onClick={() => expandField(value)}
                                    className={cn(
                                        "flex w-full items-center gap-3 rounded-lg border px-4 py-3",
                                        hasValue
                                            ? "border-border bg-background hover:bg-primary/50"
                                            : "border-border hover:border-primary/20 hover:bg-primary/50",
                                        "group text-left transition-colors",
                                    )}
                                >
                                    <Icon
                                        className={cn(
                                            "h-5 w-5",
                                            hasValue
                                                ? "text-foreground"
                                                : "text-muted-foreground group-hover:text-foreground",
                                        )}
                                        {...getIconTheme(label)}
                                    />
                                    {hasValue ? (
                                        <div className="flex flex-1 items-center gap-2">
                                            <span className="text-sm font-medium text-foreground">
                                                {label}
                                            </span>
                                            <span className="text-muted-foreground text-sm">
                                                @{displayUsername}
                                            </span>
                                        </div>
                                    ) : (
                                        <span className="text-muted-foreground text-sm font-medium transition-colors group-hover:text-foreground">
                                            {t("common:add")} {label}
                                        </span>
                                    )}
                                </button>
                            ) : (
                                <div className="flex items-center gap-3 rounded-lg border border-primary/40 bg-primary/20 px-4 py-3">
                                    <Icon
                                        className="h-5 w-5 flex-shrink-0"
                                        {...getIconTheme(label)}
                                    />
                                    <div className="relative flex-1">
                                        <Input
                                            type="url"
                                            placeholder={placeholder}
                                            value={links[value] || ""}
                                            onChange={(e) =>
                                                handleInputChange(
                                                    value,
                                                    e.target.value,
                                                )
                                            }
                                            className="h-auto w-full border-none bg-transparent p-0 pr-10 shadow-none focus-visible:ring-0 focus-visible:ring-offset-0"
                                            autoFocus={true}
                                            onKeyDown={(e) => {
                                                if (e.key === "Enter") {
                                                    e.preventDefault();
                                                    handleSave(
                                                        value,
                                                        links[value] || "",
                                                    );
                                                }
                                            }}
                                        />
                                        <button
                                            onClick={() =>
                                                handleSave(
                                                    value,
                                                    links[value] || "",
                                                )
                                            }
                                            className="absolute right-12 top-1/2 -translate-y-1/2 text-primary-foreground transition-colors"
                                            aria-label="Update field button"
                                        >
                                            <IconCheck className="h-4 w-4" />
                                        </button>
                                        <button
                                            onClick={() =>
                                                collapseField(value, !hasValue)
                                            }
                                            className="absolute right-3 top-1/2 -translate-y-1/2 text-primary-foreground transition-colors"
                                            aria-label="Close field button"
                                        >
                                            <IconClose className="h-4 w-4" />
                                        </button>
                                    </div>
                                </div>
                            )}
                        </div>
                    );
                },
            )}
        </div>
    );
}
