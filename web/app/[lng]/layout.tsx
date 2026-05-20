import { Inter } from "next/font/google";
import "@/app/globals.css";
import React, { Suspense } from "react";
import Loading from "./loading";
import Toast from "@/components/utils/toast";
import UploadManager from "@/components/upload-manager/upload-manager";
import { I18N_FALLBACK_LNG, I18N_LANGUAGES } from "@/lib/i18n/settings";
import { dir } from "i18next";
import { myGetT } from "@/lib/i18n";
import TranslationsProvider from "@/components/utils/i18n-provider";
import { ThemeProviderWrapper } from "@/components/utils/theme-provider-wrapper";
import UserInformationWrapper from "@/components/wrappers/UserInformationWrapper";
import MockProvider from "@/components/utils/mock-provider";
import type { Metadata } from "next";

const USE_MOCK_API = process.env.NEXT_PUBLIC_USE_MOCK_API === "true";
const SITE_URL =
    process.env.NEXT_PUBLIC_SITE_URL?.trim() || "http://localhost:5000";

const inter = Inter({ subsets: ["latin"] });
type Params = Promise<{ lng: string }>;

export async function generateStaticParams() {
    return I18N_LANGUAGES.map((language) => ({
        lng: language,
    }));
}

export async function generateMetadata({
    params,
}: {
    params: Params;
}): Promise<Metadata> {
    const { lng } = await params;
    const { t } = await myGetT("common");
    const title = t("app_title");
    const description = t("app_description");
    const languages = Object.fromEntries(
        I18N_LANGUAGES.map((l) => [l, `/${l}`]),
    );
    languages["x-default"] = `/${I18N_FALLBACK_LNG}`;

    return {
        metadataBase: new URL(SITE_URL),
        title: {
            default: title,
            template: `%s · ${title}`,
        },
        description,
        applicationName: title,
        alternates: {
            canonical: `/${lng}`,
            languages,
        },
        openGraph: {
            type: "website",
            siteName: title,
            title,
            description,
            url: `/${lng}`,
            locale: lng,
        },
        twitter: {
            card: "summary_large_image",
            title,
            description,
        },
        robots: {
            index: true,
            follow: true,
        },
    };
}

export default async function RootLayout({
    children,
    params,
}: {
    children: React.ReactNode;
    params: Params;
}) {
    const { lng } = await params;

    const content = (
        <TranslationsProvider>
            <ThemeProviderWrapper>
                <Suspense fallback={<Loading />}>
                    <UserInformationWrapper>{children}</UserInformationWrapper>
                    <Toast />
                    <UploadManager />
                </Suspense>
            </ThemeProviderWrapper>
        </TranslationsProvider>
    );

    return (
        <html lang={lng} dir={dir(lng)}>
            <body className={inter.className}>
                {USE_MOCK_API ? (
                    <MockProvider>{content}</MockProvider>
                ) : (
                    content
                )}
            </body>
        </html>
    );
}
