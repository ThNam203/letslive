import {
    I18N_COOKIE_MAX_AGE_SECONDS,
    I18N_COOKIE_NAME,
    I18N_FALLBACK_LNG,
    I18N_HEADER_NAME,
    I18N_LANGUAGES,
    isSupportedLocale,
    matchAcceptLanguage,
} from "@/lib/i18n/settings";
import { NextRequest, NextResponse } from "next/server";

function setLocaleCookie(response: NextResponse, locale: string) {
    response.cookies.set(I18N_COOKIE_NAME, locale, {
        path: "/",
        maxAge: I18N_COOKIE_MAX_AGE_SECONDS,
        sameSite: "lax",
    });
}

function findLegacyLocalePrefix(pathname: string): string | undefined {
    return I18N_LANGUAGES.find(
        (l) => pathname === `/${l}` || pathname.startsWith(`/${l}/`),
    );
}

/**
 * Resolves the request's locale and hands it to server components through a
 * request header. The locale is not part of the URL: the cookie (kept in sync
 * with the user's profile on the client) wins, then Accept-Language.
 */
export async function proxy(request: NextRequest) {
    const { pathname, search } = request.nextUrl;
    const cookieLocale = request.cookies.get(I18N_COOKIE_NAME)?.value;
    const hasValidCookie = isSupportedLocale(cookieLocale);

    // old links still carry the locale prefix, e.g. /vi-VN/users/1
    const legacyLocale = findLegacyLocalePrefix(pathname);
    if (legacyLocale) {
        const stripped = pathname.slice(legacyLocale.length + 1) || "/";
        const response = NextResponse.redirect(
            new URL(`${stripped}${search}`, request.url),
            308,
        );
        if (!hasValidCookie) setLocaleCookie(response, legacyLocale);
        return response;
    }

    const locale = hasValidCookie
        ? cookieLocale
        : (matchAcceptLanguage(request.headers.get("accept-language")) ??
          I18N_FALLBACK_LNG);

    const headers = new Headers(request.headers);
    headers.set(I18N_HEADER_NAME, locale);

    const response = NextResponse.next({ request: { headers } });
    if (!hasValidCookie) setLocaleCookie(response, locale);
    return response;
}

export const config = {
    matcher: [
        "/((?!api|_next/static|_next/image|favicon.ico|images|assets|png|svg|jpg|jpeg|gif|webp|mockServiceWorker\\.js).*)",
    ],
};
