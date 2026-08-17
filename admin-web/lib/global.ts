function getAdminApiUrl(): string {
    const url = process.env.NEXT_PUBLIC_ADMIN_API_URL?.trim();

    if (!url) {
        if (typeof window !== "undefined") {
            console.error("Missing NEXT_PUBLIC_ADMIN_API_URL environment variable");
        }
        throw new Error("Missing required environment variable: NEXT_PUBLIC_ADMIN_API_URL");
    }

    return url;
}

const GLOBAL = Object.freeze({
    ADMIN_API_URL: getAdminApiUrl(),
});

export default GLOBAL;
