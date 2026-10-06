import { CHAT_MESSAGE_TYPE } from "@/constant/chat";

type MessageType = (typeof CHAT_MESSAGE_TYPE)[keyof typeof CHAT_MESSAGE_TYPE];

export type SendMessage = {
    type: MessageType;
    roomId: string;
    userId: string;
    username: string;
    text: string;
};

export type ReceivedMessage = {
    id: string;
    type: MessageType;
    userId: string;
    username: string;
    text: string;
    // null when the sender has no avatar; absent on messages stored before it was added
    profilePicture?: string | null;
    // epoch ms over the socket, ISO string from the history endpoint
    timestamp: number | string;
};
