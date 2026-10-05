import { Inter } from "next/font/google";
import "@/app/globals.css";
import React, { Suspense } from "react";
import Toast from "@/components/utils/toast";
import UploadManager from "@/components/upload-manager/upload-manager";
import { dir } from "i18next";
import { getLocale, myGetT } from "@/lib/i18n";
import { getSiteUrl } from "@/utils/siteUrl";
import type { Metadata } from "next";
import TranslationsProvider from "@/components/utils/i18n-provider";
import { ThemeProviderWrapper } from "@/components/utils/theme-provider-wrapper";
import UserInformationWrapper from "@/components/wrappers/UserInformationWrapper";
import MockProvider from "@/components/utils/mock-provider";
import QueryProvider from "@/components/utils/query-provider";
import CookieConsentBanner from "@/components/utils/cookie-consent-banner";

const USE_MOCK_API = process.env.NEXT_PUBLIC_USE_MOCK_API === "true";

const inter = Inter({ subsets: ["latin"] });

/**
 * Localized site-wide defaults. Next shallowly merges a page's metadata with
 * these values; nested blocks such as openGraph are replaced as a whole.
 * A page setting a plain `title` gets it run through the template below; one
 * that needs to stand alone uses `title: { absolute: ... }`.
 *
 * The default card image is app/opengraph-image.tsx, which Next attaches to
 * any level that does not set openGraph.images itself.
 * Translation and request-header errors propagate; an invalid site URL throws
 * a TypeError when constructing metadataBase.
 */
export async function generateMetadata(): Promise<Metadata> {
    const { t, lng } = await myGetT("common");
    const siteUrl = await getSiteUrl();
    const appTitle = t("common:app_title");
    const description = t("common:app_description");

    return {
        metadataBase: new URL(siteUrl),
        title: {
            default: appTitle,
            template: `%s | ${appTitle}`,
        },
        description,
        applicationName: appTitle,
        robots: { index: true, follow: true },
        openGraph: {
            type: "website",
            siteName: appTitle,
            title: appTitle,
            description,
            locale: lng.replace("-", "_"),
        },
        twitter: {
            card: "summary_large_image",
            title: appTitle,
            description,
        },
    };
}

export default async function RootLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    const lng = await getLocale();

    const content = (
        <ThemeProviderWrapper>
            {/* no boundary above the page: once a fallback streams, the
                status is fixed at 200 and notFound() cannot send 404.
                Segments own their loading.tsx and skeletons instead. */}
            <UserInformationWrapper>{children}</UserInformationWrapper>
            <Suspense fallback={null}>
                <Toast />
                <UploadManager />
                <CookieConsentBanner />
            </Suspense>
        </ThemeProviderWrapper>
    );

    return (
        <html lang={lng} dir={dir(lng)}>
            <body className={inter.className}>
                <QueryProvider>
                    <TranslationsProvider lng={lng}>
                        {USE_MOCK_API ? (
                            <MockProvider>{content}</MockProvider>
                        ) : (
                            content
                        )}
                    </TranslationsProvider>
                </QueryProvider>
            </body>
        </html>
    );
}
