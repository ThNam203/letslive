import { cache } from "react";
import type { Metadata } from "next";
import { GetVODInformation } from "@/lib/api/vod";
import { GetUserById } from "@/lib/api/user";
import { myGetT } from "@/lib/i18n";
import { I18N_FALLBACK_LNG, I18N_LANGUAGES } from "@/lib/i18n/settings";
import { JsonLd } from "@/components/seo/json-ld";
import type { VOD } from "@/types/vod";
import type { PublicUser } from "@/types/user";

const SITE_URL =
    process.env.NEXT_PUBLIC_SITE_URL?.trim() || "http://localhost:5000";

type Params = Promise<{ lng: string; userId: string; vodId: string }>;

const loadVod = cache(async (vodId: string, userId: string) => {
    let vod: VOD | null = null;
    let user: PublicUser | null = null;
    try {
        const [vodRes, userRes] = await Promise.all([
            GetVODInformation(vodId),
            GetUserById(userId).catch(() => null),
        ]);
        if (vodRes.success && vodRes.data) vod = vodRes.data;
        if (userRes && userRes.success && userRes.data) user = userRes.data;
    } catch (err) {
        console.error("vod layout fetch failed", err);
    }
    return { vod, user };
});

function isoDuration(seconds: number | null | undefined): string | undefined {
    if (!seconds || seconds <= 0) return undefined;
    const s = Math.floor(seconds);
    const h = Math.floor(s / 3600);
    const m = Math.floor((s % 3600) / 60);
    const sec = s % 60;
    return `PT${h}H${m}M${sec}S`;
}

export async function generateMetadata({
    params,
}: {
    params: Params;
}): Promise<Metadata> {
    const { lng, userId, vodId } = await params;
    const { t } = await myGetT("common");
    const { vod, user } = await loadVod(vodId, userId);

    if (!vod || vod.visibility !== "public" || vod.status !== "ready") {
        return { robots: { index: false, follow: false } };
    }

    const username = user?.username || "";
    const title = vod.title;
    const description = vod.description
        ? vod.description
        : t("meta_vod_description", { title: vod.title, username });
    const path = `/users/${userId}/vods/${vodId}`;
    const languages = Object.fromEntries(
        I18N_LANGUAGES.map((l) => [l, `/${l}${path}`]),
    );
    languages["x-default"] = `/${I18N_FALLBACK_LNG}${path}`;
    const images = vod.thumbnailUrl ? [{ url: vod.thumbnailUrl }] : undefined;

    return {
        title,
        description,
        alternates: { canonical: `/${lng}${path}`, languages },
        openGraph: {
            type: "video.other",
            title,
            description,
            url: `/${lng}${path}`,
            images,
        },
        twitter: {
            card: "summary_large_image",
            title,
            description,
            images: vod.thumbnailUrl ? [vod.thumbnailUrl] : undefined,
        },
    };
}

export default async function VODLayout({
    children,
    params,
}: {
    children: React.ReactNode;
    params: Params;
}) {
    const { lng, userId, vodId } = await params;
    const { vod, user } = await loadVod(vodId, userId);

    if (!vod || vod.visibility !== "public" || vod.status !== "ready") {
        return <>{children}</>;
    }

    const pageUrl = `${SITE_URL}/${lng}/users/${userId}/vods/${vodId}`;
    const videoObject: Record<string, unknown> = {
        "@context": "https://schema.org",
        "@type": "VideoObject",
        name: vod.title,
        description: vod.description ?? undefined,
        uploadDate: vod.createdAt,
        url: pageUrl,
        thumbnailUrl: vod.thumbnailUrl ?? undefined,
        contentUrl: vod.playbackUrl ?? undefined,
        duration: isoDuration(vod.duration),
        interactionStatistic: {
            "@type": "InteractionCounter",
            interactionType: { "@type": "WatchAction" },
            userInteractionCount: vod.viewCount,
        },
    };
    if (user?.username) {
        videoObject.author = {
            "@type": "Person",
            name: user.username,
            url: `${SITE_URL}/${lng}/users/${userId}`,
        };
    }

    return (
        <>
            <JsonLd data={videoObject} />
            {children}
        </>
    );
}
