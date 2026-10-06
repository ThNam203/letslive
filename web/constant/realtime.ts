export const REALTIME_EVENT = {
    NOTIFICATION_CREATED: "notification.created",
    CHAT_MESSAGE: "chat.message",
    MEMBER_JOINED: "member.joined",
    MEMBER_LEFT: "member.left",
} as const;

export const realtimeRoomTopic = (roomId: string) => `room:${roomId}`;

export const REALTIME_RECONNECT_INITIAL_DELAY_MS = 1000;
export const REALTIME_RECONNECT_MAX_DELAY_MS = 30_000;
