function getBackendUrl(): string {
    const protocol = process.env.NEXT_PUBLIC_BACKEND_PROTOCOL?.trim();
    const ipAddress = process.env.NEXT_PUBLIC_BACKEND_IP_ADDRESS?.trim();
    const port = process.env.NEXT_PUBLIC_BACKEND_PORT?.trim();

    if (
        !protocol ||
        !ipAddress ||
        !port ||
        protocol === "" ||
        ipAddress === "" ||
        port === ""
    ) {
        if (typeof window !== "undefined") {
            console.error("Missing or empty backend environment variables:", {
                protocol: protocol || "(empty or undefined)",
                ipAddress: ipAddress || "(empty or undefined)",
                port: port || "(empty or undefined)",
            });
        }
        throw new Error(
            "Missing or empty required environment variables: NEXT_PUBLIC_BACKEND_PROTOCOL, NEXT_PUBLIC_BACKEND_IP_ADDRESS, NEXT_PUBLIC_BACKEND_PORT",
        );
    }

    return `${protocol}://${ipAddress}:${port}`;
}

function getWebSocketUrl(): string {
    const wsProtocol = process.env.NEXT_PUBLIC_BACKEND_WS_PROTOCOL?.trim();
    const ipAddress = process.env.NEXT_PUBLIC_BACKEND_IP_ADDRESS?.trim();
    const port = process.env.NEXT_PUBLIC_BACKEND_PORT?.trim();

    if (
        !wsProtocol ||
        !ipAddress ||
        !port ||
        wsProtocol === "" ||
        ipAddress === "" ||
        port === ""
    ) {
        if (typeof window !== "undefined") {
            console.error("Missing or empty WebSocket environment variables:", {
                wsProtocol: wsProtocol || "(empty or undefined)",
                ipAddress: ipAddress || "(empty or undefined)",
                port: port || "(empty or undefined)",
            });
        }
        throw new Error(
            "Missing or empty required environment variables: NEXT_PUBLIC_BACKEND_WS_PROTOCOL, NEXT_PUBLIC_BACKEND_IP_ADDRESS, NEXT_PUBLIC_BACKEND_PORT",
        );
    }

    return `${wsProtocol}://${ipAddress}:${port}/ws`;
}

const GLOBAL = Object.freeze({
    API_URL: getBackendUrl(),
    WS_SERVER_URL: getWebSocketUrl(),
});

export default GLOBAL;
