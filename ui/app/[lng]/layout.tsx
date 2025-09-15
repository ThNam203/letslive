import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "@/app/globals.css";
import React, { Suspense } from "react";
import Loading from "./loading";
import Toast from "@/components/utils/toast";
import { ThemeProvider } from 'next-themes'

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
    title: "Let's Live",
    description: "A platform for live streaming",
};

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <html lang="en" suppressHydrationWarning>
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
