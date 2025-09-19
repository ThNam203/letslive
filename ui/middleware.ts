import {
    I18N_COOKIE_NAME,
    I18N_FALLBACK_LNG,
    I18N_HEADER_NAME,
    I18N_LANGUAGES,
} from "@/lib/i18n/settings";
import { NextRequest, NextResponse } from "next/server";

function setUpCookieLocale(
    request: NextRequest,
    response: NextResponse,
): string {
    // check cookie
    const localeCookie = request.cookies.get(I18N_COOKIE_NAME)?.value;
    if (localeCookie) return localeCookie;

    let finalLng = I18N_FALLBACK_LNG;

    // then check Accept-Language
    const acceptLang = request.headers.get("accept-language")?.split(",")[0];
    if (acceptLang)
        finalLng = acceptLang.split("-")[0]; // just "en" if "en-US"
    else if (request.headers.has("referer")) {
        const refererUrl = new URL(request.headers.get("referer") || "");
        finalLng =
            I18N_LANGUAGES.find((l) => refererUrl.pathname.startsWith(`/${l}`)) ||
            "";
    }

    if (!I18N_LANGUAGES.includes(finalLng)) finalLng = I18N_FALLBACK_LNG;

    response.cookies.set(I18N_COOKIE_NAME, finalLng, { maxAge: 60 * 60 * 24 * 365 });
    return finalLng;
}

/**
 * Middleware to handle locale selection.
 */
export async function middleware(request: NextRequest) {
    let redirectUrl = request.nextUrl.href;
    let response = NextResponse.redirect(redirectUrl, 307);

    const deprivedLocale = setUpCookieLocale(request, response);
    const localeInPath = I18N_LANGUAGES.find((loc) =>
        request.nextUrl.pathname.startsWith(`/${loc}`),
    );

    // if locale is already in pathname, use it
    // if not then use locale from cookie or accept-language header
    const locale = localeInPath || deprivedLocale;

    const headers = new Headers(request.headers);
    headers.set(I18N_HEADER_NAME, locale);

    if (!localeInPath && !request.nextUrl.pathname.startsWith("/_next")) {
        return NextResponse.redirect(
            new URL(
                `/${locale}${request.nextUrl.pathname}${request.nextUrl.search}`,
                request.url,
            ),
        );
    }

    // check if the url is a static asset
    if (request.nextUrl.pathname.includes(".")) {
        return NextResponse.next({ headers });
    }

    return NextResponse.next({ headers });
}

export const config = {
    matcher: [
        "/((?!api|_next/static|_next/image|favicon.ico|images|assets|png|svg|jpg|jpeg|gif|webp).*)",
    ],
};
