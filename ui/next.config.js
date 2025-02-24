module.exports = {
    reactStrictMode: false,
    images: {
        domains: ["github.com", "placehold.co", "localhost", "minio"],
        remotePatterns: [
            {
                protocol: "https",
                hostname: "**",
            },
        ],
    },
};
