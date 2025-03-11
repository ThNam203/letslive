const GLOBAL = Object.freeze({
    API_URL: process.env.NEXT_PUBLIC_BACKEND_PROTOCOL + "://" + process.env.NEXT_PUBLIC_BACKEND_IP_ADDRESS + ":" + process.env.NEXT_PUBLIC_BACKEND_PORT,
    WS_SERVER_URL: process.env.NEXT_PUBLIC_BACKEND_WS_PROTOCOL + "://" + process.env.NEXT_PUBLIC_BACKEND_IP_ADDRESS + ":" + process.env.NEXT_PUBLIC_BACKEND_PORT + "/ws",
})

export default GLOBAL;