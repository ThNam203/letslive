import { headers } from "next/headers";

/**
 * Absolute origin of the web app, for metadata that must not be relative:
 * link previews in chat apps are rendered by crawlers that never resolve a
 * path against the page they fetched.
 *
 * NEXT_PUBLIC_SITE_URL is the source of truth; behind a proxy we fall back to
 * the forwarded host, and to localhost in development.
 */
export async function getSiteUrl(): Promise<string> {
    const configured = process.env.NEXT_PUBLIC_SITE_URL?.trim();
    if (configured) {
        return configured.replace(/\/+$/, "");
    }

    const headerList = await headers();
    const host =
        headerList.get("x-forwarded-host") ?? headerList.get("host") ?? null;
    if (host) {
        const protocol =
            headerList.get("x-forwarded-proto") ??
            (host.startsWith("localhost") ? "http" : "https");
        return `${protocol}://${host}`;
    }

    return `http://localhost:${process.env.PORT ?? 3000}`;
}

/** Makes a possibly site-relative URL absolute; drops anything unusable. */
export function toAbsoluteUrl(
    url: string | null | undefined,
    siteUrl: string,
): string | null {
    if (!url) return null;
    try {
        return new URL(url, siteUrl).toString();
    } catch {
        return null;
    }
}
