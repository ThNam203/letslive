import {
    cookieName,
    fallbackLng,
    headerName,
    languages,
} from "@/lib/i18n/settings";
import { NextRequest, NextResponse } from "next/server";

function setUpCookieLocale(
    request: NextRequest,
    response: NextResponse,
): string {
    // check cookie
    const localeCookie = request.cookies.get(cookieName)?.value;
    if (localeCookie) return localeCookie;

    let finalLng = fallbackLng;

    // then check Accept-Language
    const acceptLang = request.headers.get("accept-language")?.split(",")[0];
    if (acceptLang)
        finalLng = acceptLang.split("-")[0]; // just "en" if "en-US"
    else if (request.headers.has("referer")) {
        const refererUrl = new URL(request.headers.get("referer") || "");
        finalLng =
            languages.find((l) => refererUrl.pathname.startsWith(`/${l}`)) ||
            "";
    }

    if (!languages.includes(finalLng)) finalLng = fallbackLng;

    response.cookies.set(cookieName, finalLng, { maxAge: 60 * 60 * 24 * 365 });
    return finalLng;
}

/**
 * Middleware to handle locale selection.
 */
export async function middleware(request: NextRequest) {
    let redirectUrl = request.nextUrl.href;
    let response = NextResponse.redirect(redirectUrl, 307);

    const deprivedLocale = setUpCookieLocale(request, response);
    const localeInPath = languages.find((loc) =>
        request.nextUrl.pathname.startsWith(`/${loc}`),
    );

    // if locale is already in pathname, use it
    // if not then use locale from cookie or accept-language header
    const locale = localeInPath || deprivedLocale;

    const headers = new Headers(request.headers);
    headers.set(headerName, locale);

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
