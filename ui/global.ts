const GLOBAL = Object.freeze({
    API_URL: "http://" + process.env.NEXT_PUBLIC_BACKEND_IP_ADDRESS + ":" + process.env.NEXT_PUBLIC_BACKEND_PORT,
    WS_SERVER_URL: "ws://" + process.env.NEXT_PUBLIC_BACKEND_IP_ADDRESS + ":" + process.env.NEXT_PUBLIC_BACKEND_PORT + "/ws",
})

export default GLOBAL;