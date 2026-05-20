import type { MetadataRoute } from "next";

const SITE_URL =
    process.env.NEXT_PUBLIC_SITE_URL?.trim() || "http://localhost:5000";

export default function robots(): MetadataRoute.Robots {
    return {
        rules: [
            {
                userAgent: "*",
                allow: "/",
                disallow: [
                    "/api/",
                    "/*/login",
                    "/*/signup",
                    "/*/account-setup",
                    "/*/settings",
                    "/*/messages",
                    "/*/notifications",
                    "/*/wallet",
                ],
            },
        ],
        sitemap: `${SITE_URL}/sitemap.xml`,
        host: SITE_URL,
    };
}
