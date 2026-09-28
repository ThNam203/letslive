import { Inter } from "next/font/google";
import "@/app/globals.css";
import React, { Suspense } from "react";
import Loading from "./loading";
import Toast from "@/components/utils/toast";
import UploadManager from "@/components/upload-manager/upload-manager";
import { dir } from "i18next";
import { getLocale, myGetT } from "@/lib/i18n";
import TranslationsProvider from "@/components/utils/i18n-provider";
import { ThemeProviderWrapper } from "@/components/utils/theme-provider-wrapper";
import UserInformationWrapper from "@/components/wrappers/UserInformationWrapper";
import MockProvider from "@/components/utils/mock-provider";
import QueryProvider from "@/components/utils/query-provider";
import CookieConsentBanner from "@/components/utils/cookie-consent-banner";

const USE_MOCK_API = process.env.NEXT_PUBLIC_USE_MOCK_API === "true";

const inter = Inter({ subsets: ["latin"] });

export async function generateMetadata() {
    const { t } = await myGetT("common");

    return {
        title: t("app_title"),
    };
}

export default async function RootLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    const lng = await getLocale();

    const content = (
        <QueryProvider>
            <TranslationsProvider lng={lng}>
                <ThemeProviderWrapper>
                    <Suspense fallback={<Loading />}>
                        <UserInformationWrapper>
                            {children}
                        </UserInformationWrapper>
                        <Toast />
                        <UploadManager />
                        <CookieConsentBanner />
                    </Suspense>
                </ThemeProviderWrapper>
            </TranslationsProvider>
        </QueryProvider>
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
