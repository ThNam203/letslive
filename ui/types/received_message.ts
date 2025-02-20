export type ReceivedMessage = {
    id: string;
    type: "message" | "join" | "leave";
    userId: string;
    username: string;
    text: string;
    timestamp: number;
};