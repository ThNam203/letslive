// A live chat line: a history item from GET /messages or a chat.message push.
export type ReceivedMessage = {
    userId: string;
    username: string;
    text: string;
    // null when the sender has no avatar; absent on messages stored before it was added
    profilePicture?: string | null;
    // epoch ms in pushes, ISO string from the history endpoint
    timestamp: number | string;
};

export type ChatMemberEvent = {
    userId: string;
};
