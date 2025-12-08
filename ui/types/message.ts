export type SendMessage = {
    type: "message" | "join" | "leave";
    roomId: string;
    userId: string;
    username: string;
    text: string;
};

export type ReceivedMessage = {
    id: string;
    type: "message" | "join" | "leave";
    userId: string;
    username: string;
    text: string;
    timestamp: number;
};
