module.exports = {
    reactStrictMode: false,
    images: {
        domains: ["github.com", "placehold.co", "localhost", "minio", "kong"],
        remotePatterns: [
            {
                protocol: "https",
                hostname: "**",
            },
        ],
    },
};
