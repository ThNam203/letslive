import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { NextUIProvider } from "@nextui-org/system";
import StoreProvider from "@/redux/provider";
import Toast from "@/components/Toaster";
import React, { Suspense } from "react";
import Loading from "@/app/loading";
import { HeliaProvider } from "@/helia/provider";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
    title: "Let's Live",
    description: "Powered by NextJS",
};

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <html lang="en">
            <body className={inter.className}>
                <StoreProvider>
                    <HeliaProvider>
                    <Suspense fallback={<Loading />}>
                        <NextUIProvider>{children}</NextUIProvider>
                        <Toast />
                    </Suspense>
                    </HeliaProvider>
                </StoreProvider>
            </body>
        </html>
    );
}
