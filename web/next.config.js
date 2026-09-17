const { withSentryConfig } = require("@sentry/nextjs");

/** @type {import('next').NextConfig} */
const nextConfig = {
    reactStrictMode: false,
    // Vercel runs its own output file tracing; "standalone" removes the
    // .nft.json files its build adapter expects and fails onBuildComplete.
    output: process.env.VERCEL ? undefined : "standalone",
    images: {
        formats: ["image/avif", "image/webp"],
        remotePatterns: [
            {
                protocol: "https",
                hostname: "*",
            },
            {
                protocol: "http",
                hostname: "localhost",
            },
        ],
    },
    experimental: {
        optimizePackageImports: [
            "@radix-ui/react-avatar",
            "@radix-ui/react-dialog",
            "@radix-ui/react-dropdown-menu",
            "@radix-ui/react-hover-card",
            "@radix-ui/react-label",
            "@radix-ui/react-popover",
            "@radix-ui/react-scroll-area",
            "@radix-ui/react-select",
            "@radix-ui/react-separator",
            "@radix-ui/react-slider",
            "@radix-ui/react-slot",
            "@radix-ui/react-switch",
            "@radix-ui/react-tabs",
            "@radix-ui/react-tooltip",
            "lucide-react",
            "@sentry/nextjs",
        ],
    },
};

const isProductionEnvironment =
    process.env.NEXT_PUBLIC_ENVIRONMENT === "production";

const sentryWebpackPluginOptions = {
    org: "vnuhcm-uit-university-of-infor",
    project: "letslive",
    silent: !process.env.CI,
    authToken: process.env.SENTRY_AUTH_TOKEN,
    widenClientFileUpload: true,
};

module.exports = isProductionEnvironment
    ? withSentryConfig(nextConfig, sentryWebpackPluginOptions)
    : nextConfig;
