import { Inter } from "next/font/google";
import "@/app/globals.css";
import React, { Suspense } from "react";
import Loading from "./loading";
import Toast from "@/components/utils/toast";
import { I18N_LANGUAGES } from "@/lib/i18n/settings";
import { dir } from "i18next";
import { myGetT } from "@/lib/i18n";
import TranslationsProvider from "@/components/utils/i18n-provider";
import { ThemeProviderWrapper } from "@/components/utils/theme-provider-wrapper";

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
    
    return (
        <html lang={lng} dir={dir(lng)}>
            <body className={inter.className}>
                <TranslationsProvider>
                    <ThemeProviderWrapper>
                        <Suspense fallback={<Loading />}>
                            {children}
                            <Toast />
                        </Suspense>
                    </ThemeProviderWrapper>
                </TranslationsProvider>
            </body>
        </html>
    );
}
