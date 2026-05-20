import type { Metadata } from "next";
import { GetVODInformation } from "@/lib/api/vod";
import { GetUserById } from "@/lib/api/user";
import { myGetT } from "@/lib/i18n";
import { I18N_FALLBACK_LNG, I18N_LANGUAGES } from "@/lib/i18n/settings";

type Params = Promise<{ lng: string; userId: string; vodId: string }>;

export async function generateMetadata({
    params,
}: {
    params: Params;
}): Promise<Metadata> {
    const { lng, userId, vodId } = await params;
    const { t } = await myGetT("common");

    try {
        const [vodRes, userRes] = await Promise.all([
            GetVODInformation(vodId),
            GetUserById(userId).catch(() => null),
        ]);

        if (!vodRes.success || !vodRes.data) {
            return { robots: { index: false, follow: false } };
        }

        const vod = vodRes.data;
        if (vod.visibility !== "public" || vod.status !== "ready") {
            return { robots: { index: false, follow: false } };
        }

        const username =
            userRes && userRes.success && userRes.data
                ? userRes.data.username
                : "";

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
    } catch (err) {
        console.error("vod metadata fetch failed", err);
        return { robots: { index: false, follow: false } };
    }
}

export default function VODLayout({ children }: { children: React.ReactNode }) {
    return children;
}
