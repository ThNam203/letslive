export type ConversationParticipant = {
    userId: string;
    username: string;
    displayName: string | null;
    profilePicture: string | null;
    role: "owner" | "admin" | "member";
    joinedAt: string;
    lastReadMessageId: string | null;
    isMuted: boolean;
};

export type LastMessage = {
    _id: string;
    senderId: string;
    senderUsername: string;
    text: string;
    createdAt: string;
};

export type Conversation = {
    _id: string;
    type: "dm" | "group";
    name: string | null;
    avatarUrl: string | null;
    createdBy: string;
    participants: ConversationParticipant[];
    lastMessage: LastMessage | null;
    createdAt: string;
    updatedAt: string;
};

export type ReadReceipt = {
    userId: string;
    readAt: string;
};

export type DmMessage = {
    _id: string;
    conversationId: string;
    senderId: string;
    senderUsername: string;
    type: "text" | "image" | "system";
    text: string;
    imageUrls?: string[];
    replyTo?: string;
    isDeleted: boolean;
    readBy: ReadReceipt[];
    createdAt: string;
    updatedAt: string;
};

// WebSocket event types (client → server)
export type DmWsSendMessage = {
    type: "dm:send_message";
    conversationId: string;
    text: string;
    messageType: "text" | "image";
    senderUsername: string;
    imageUrls?: string[];
    replyTo?: string;
};

export type DmWsTyping = {
    type: "dm:typing_start" | "dm:typing_stop";
    conversationId: string;
    username: string;
};

export type DmWsMarkRead = {
    type: "dm:mark_read";
    conversationId: string;
    messageId: string;
};

export type DmWsClientEvent =
    | DmWsSendMessage
    | DmWsTyping
    | DmWsMarkRead;

// WebSocket event types (server → client)
export type DmWsNewMessage = {
    type: "dm:new_message";
    conversationId: string;
    message: DmMessage;
};

export type DmWsMessageEdited = {
    type: "dm:message_edited";
    conversationId: string;
    messageId: string;
    newText: string;
    updatedAt: string;
};

export type DmWsMessageDeleted = {
    type: "dm:message_deleted";
    conversationId: string;
    messageId: string;
};

export type DmWsUserTyping = {
    type: "dm:user_typing" | "dm:user_stopped_typing";
    conversationId: string;
    userId: string;
    username: string;
};

export type DmWsReadReceipt = {
    type: "dm:read_receipt";
    conversationId: string;
    userId: string;
    messageId: string;
    readAt: string;
};

export type DmWsPresence = {
    type: "dm:user_online" | "dm:user_offline";
    userId: string;
};

export type DmWsConversationUpdated = {
    type: "dm:conversation_updated";
    conversationId: string;
    update: Partial<Conversation>;
};

export type DmWsServerEvent =
    | DmWsNewMessage
    | DmWsMessageEdited
    | DmWsMessageDeleted
    | DmWsUserTyping
    | DmWsReadReceipt
    | DmWsPresence
    | DmWsConversationUpdated;
