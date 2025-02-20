export type SendMessage = {
    type: "message" | "join" | "leave";
    roomId: string;
    userId: string;
    username: string;
    text: string;
};