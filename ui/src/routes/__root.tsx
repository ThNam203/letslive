import type { ReactNode } from "react";
import {
    Outlet,
    createRootRoute,
    HeadContent,
    Scripts,
} from "@tanstack/react-router";
import "@/routes/globals.css";
import TranslationsProvider from "@/components/utils/i18n-provider";
import { ThemeProviderWrapper } from "@/components/utils/theme-provider-wrapper";

export const Route = createRootRoute({
    head: () => ({
        meta: [
            { charSet: "utf-8" },
            {
                name: "viewport",
                content: "width=device-width, initial-scale=1",
            },
            { title: "Let's Live" },
        ],
    }),
    component: RootComponent,
});

function RootComponent() {
    return (
        <RootDocument>
            <TranslationsProvider>
                <ThemeProviderWrapper>
                    <Outlet />
                </ThemeProviderWrapper>
            </TranslationsProvider>
        </RootDocument>
    );
}

function RootDocument({ children }: Readonly<{ children: ReactNode }>) {
    return (
        <html>
            <head>
                <HeadContent />
            </head>
            <body className="font-sans antialiased">
                {children}
                <Scripts />
            </body>
        </html>
    );
}
