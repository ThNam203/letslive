import { headers } from "next/headers";

/**
 * Returns the web app's base URL for metadata. A nonblank NEXT_PUBLIC_SITE_URL
 * is trimmed and returned without trailing slashes or URL validation.
 *
 * Otherwise uses x-forwarded-host, then host, with x-forwarded-proto or a
 * default of HTTP for hosts starting with localhost and HTTPS for others.
 * Without a host, returns http://localhost:${PORT}, defaulting to port 3000.
 * Request-header errors propagate rather than returning the localhost fallback.
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

/**
 * Resolves a URL against siteUrl and returns its serialized absolute form.
 * Returns null for empty or absent input or URL parsing errors. Accepted URL
 * schemes are not restricted to HTTP(S).
 */
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
