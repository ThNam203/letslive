import { cache } from "react";
import type { Metadata } from "next";
import { GetUserById } from "@/lib/api/user";
import { GetLivestreamOfUser } from "@/lib/api/livestream";
import { myGetT } from "@/lib/i18n";
import { I18N_FALLBACK_LNG, I18N_LANGUAGES } from "@/lib/i18n/settings";
import { JsonLd } from "@/components/seo/json-ld";
import type { PublicUser } from "@/types/user";
import type { Livestream } from "@/types/livestream";

const SITE_URL =
    process.env.NEXT_PUBLIC_SITE_URL?.trim() || "http://localhost:5000";

type Params = Promise<{ lng: string; userId: string }>;

const loadUser = cache(async (userId: string) => {
    let user: PublicUser | null = null;
    let livestream: Livestream | null = null;
    try {
        const [userRes, liveRes] = await Promise.all([
            GetUserById(userId),
            GetLivestreamOfUser(userId).catch(() => null),
        ]);
        if (userRes.success && userRes.data) user = userRes.data;
        if (liveRes && liveRes.success && liveRes.data)
            livestream = liveRes.data;
    } catch (err) {
        console.error("user layout fetch failed", err);
    }
    return { user, livestream };
});

export async function generateMetadata({
    params,
}: {
    params: Params;
}): Promise<Metadata> {
    const { lng, userId } = await params;
    const { t } = await myGetT("common");
    const { user, livestream } = await loadUser(userId);

    if (!user || !user.username) {
        return { robots: { index: false, follow: false } };
    }

    const username = user.username;
    const bio = user.bio || "";
    const avatar = user.profilePicture;
    const liveTitle = livestream?.title || null;

    const title = liveTitle
        ? t("meta_user_live", { username, title: liveTitle })
        : t("meta_user_offline", { username });
    const description = t("meta_user_description", { username, bio });
    const path = `/users/${userId}`;
    const languages = Object.fromEntries(
        I18N_LANGUAGES.map((l) => [l, `/${l}${path}`]),
    );
    languages["x-default"] = `/${I18N_FALLBACK_LNG}${path}`;

    return {
        title,
        description,
        alternates: { canonical: `/${lng}${path}`, languages },
        openGraph: {
            type: "profile",
            title,
            description,
            url: `/${lng}${path}`,
            images: avatar ? [{ url: avatar }] : undefined,
        },
        twitter: {
            card: "summary_large_image",
            title,
            description,
            images: avatar ? [avatar] : undefined,
        },
    };
}

export default async function UserLayout({
    children,
    params,
}: {
    children: React.ReactNode;
    params: Params;
}) {
    const { lng, userId } = await params;
    const { user, livestream } = await loadUser(userId);

    const profileUrl = `${SITE_URL}/${lng}/users/${userId}`;
    const jsonLd: Record<string, unknown>[] = [];

    if (user && user.username) {
        const person: Record<string, unknown> = {
            "@context": "https://schema.org",
            "@type": "Person",
            name: user.username,
            url: profileUrl,
            identifier: user.id,
        };
        if (user.profilePicture) person.image = user.profilePicture;
        if (user.bio) person.description = user.bio;

        const sameAs: string[] = [];
        const links = user.socialMediaLinks;
        if (links) {
            for (const key of [
                "facebook",
                "twitter",
                "instagram",
                "linkedin",
                "github",
                "youtube",
                "website",
                "tiktok",
            ] as const) {
                const v = links[key];
                if (typeof v === "string" && v) sameAs.push(v);
            }
        }
        if (sameAs.length > 0) person.sameAs = sameAs;

        jsonLd.push({
            "@context": "https://schema.org",
            "@type": "ProfilePage",
            url: profileUrl,
            mainEntity: person,
        });

        if (livestream && livestream.visibility === "public") {
            jsonLd.push({
                "@context": "https://schema.org",
                "@type": "BroadcastEvent",
                name: livestream.title,
                description: livestream.description ?? undefined,
                isLiveBroadcast: !livestream.endedAt,
                startDate: livestream.startedAt,
                endDate: livestream.endedAt ?? undefined,
                url: profileUrl,
                publishedOn: {
                    "@type": "BroadcastService",
                    name: "Let's Live",
                },
                videoFormat: "HLS",
            });
        }
    }

    return (
        <>
            {jsonLd.length > 0 && <JsonLd data={jsonLd} />}
            {children}
        </>
    );
}
