import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "@/app/globals.css";
import React, { Suspense } from "react";
import Loading from "./loading";
import Toast from "@/components/utils/toast";
import { ThemeProvider } from 'next-themes'
import { languages } from "@/lib/i18n/settings";
import { dir } from "i18next";
import { myGetT } from "@/lib/i18n";

const inter = Inter({ subsets: ["latin"] });

export async function generateStaticParams() {
    return languages.map((language) => ({
        lng: language,
    }))
}

export async function generateMetadata() {
    const { t } = await myGetT('second-page')
    return {
      title: t('title')
    }
}

export default function RootLayout({
    children,
    params,
}: Readonly<{
    children: React.ReactNode;
    params: { lng: string };
}>) {
    return (
        <html lang={params.lng} dir={dir(params.lng)} suppressHydrationWarning>
            <body className={inter.className}>
                <Suspense fallback={<Loading />}>
                    <ThemeProvider       
                        attribute="data-theme"
                        defaultTheme="system"
                        enableSystem
                    >
                        {children}
                        <Toast />
                    </ThemeProvider>
                </Suspense>
            </body>
        </html>
    );
}
