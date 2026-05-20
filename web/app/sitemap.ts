import type { MetadataRoute } from "next";
import { I18N_FALLBACK_LNG, I18N_LANGUAGES } from "@/lib/i18n/settings";
import { GetRecommendedChannels } from "@/lib/api/user";
import { GetPopularVODs } from "@/lib/api/vod";

const SITE_URL =
    process.env.NEXT_PUBLIC_SITE_URL?.trim() || "http://localhost:5000";

export const revalidate = 3600;

type SitemapEntry = MetadataRoute.Sitemap[number];

function alternates(path: string): SitemapEntry["alternates"] {
    return {
        languages: Object.fromEntries(
            I18N_LANGUAGES.map((lng) => [lng, `${SITE_URL}/${lng}${path}`]),
        ),
    };
}

function staticEntries(): MetadataRoute.Sitemap {
    const paths: Array<{ path: string; priority: number }> = [
        { path: "", priority: 1.0 },
    ];

    return paths.map(({ path, priority }) => ({
        url: `${SITE_URL}/${I18N_FALLBACK_LNG}${path}`,
        lastModified: new Date(),
        changeFrequency: "daily" as const,
        priority,
        alternates: alternates(path),
    }));
}

async function dynamicEntries(): Promise<MetadataRoute.Sitemap> {
    const entries: MetadataRoute.Sitemap = [];

    try {
        const channels = await GetRecommendedChannels(0);
        if (channels.success && channels.data) {
            for (const u of channels.data) {
                if (!u.username) continue;
                const path = `/users/${u.id}`;
                entries.push({
                    url: `${SITE_URL}/${I18N_FALLBACK_LNG}${path}`,
                    lastModified: new Date(),
                    changeFrequency: "hourly",
                    priority: 0.8,
                    alternates: alternates(path),
                });
            }
        }
    } catch (err) {
        console.error("sitemap: failed to fetch channels", err);
    }

    try {
        const vods = await GetPopularVODs(0, 100);
        if (vods.success && vods.data) {
            for (const v of vods.data) {
                if (v.visibility !== "public" || v.status !== "ready") continue;
                const path = `/users/${v.userId}/vods/${v.id}`;
                entries.push({
                    url: `${SITE_URL}/${I18N_FALLBACK_LNG}${path}`,
                    lastModified: v.updatedAt
                        ? new Date(v.updatedAt)
                        : new Date(),
                    changeFrequency: "weekly",
                    priority: 0.6,
                    alternates: alternates(path),
                });
            }
        }
    } catch (err) {
        console.error("sitemap: failed to fetch vods", err);
    }

    return entries;
}

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
    const [statics, dynamics] = await Promise.all([
        Promise.resolve(staticEntries()),
        dynamicEntries(),
    ]);
    return [...statics, ...dynamics];
}
