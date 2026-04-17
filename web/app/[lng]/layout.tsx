import { Inter } from "next/font/google";
import "@/app/globals.css";
import React, { Suspense } from "react";
import Loading from "./loading";
import Toast from "@/components/utils/toast";
import UploadManager from "@/components/upload-manager/upload-manager";
import { I18N_LANGUAGES } from "@/lib/i18n/settings";
import { dir } from "i18next";
import { myGetT } from "@/lib/i18n";
import TranslationsProvider from "@/components/utils/i18n-provider";
import { ThemeProviderWrapper } from "@/components/utils/theme-provider-wrapper";
import UserInformationWrapper from "@/components/wrappers/UserInformationWrapper";
import MockProvider from "@/components/utils/mock-provider";

const USE_MOCK_API = process.env.NEXT_PUBLIC_USE_MOCK_API === "true";

const inter = Inter({ subsets: ["latin"] });
type Params = Promise<{ lng: string }>;

export async function generateStaticParams() {
    return I18N_LANGUAGES.map((language) => ({
        lng: language,
    }));
}

export async function generateMetadata() {
    const { t } = await myGetT("common");

    return {
        title: t("app_title"),
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
                    <UserInformationWrapper>
                        {children}
                    </UserInformationWrapper>
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
