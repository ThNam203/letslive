export enum ConversationType {
    DM = "dm",
    GROUP = "group",
}

export enum DmMessageType {
    TEXT = "text",
    IMAGE = "image",
    SYSTEM = "system",
}

export enum ParticipantRole {
    OWNER = "owner",
    ADMIN = "admin",
    MEMBER = "member",
}

export enum DmClientEventType {
    SEND_MESSAGE = "dm:send_message",
    TYPING_START = "dm:typing_start",
    TYPING_STOP = "dm:typing_stop",
    MARK_READ = "dm:mark_read",
}

export enum DmServerEventType {
    NEW_MESSAGE = "dm:new_message",
    MESSAGE_EDITED = "dm:message_edited",
    MESSAGE_DELETED = "dm:message_deleted",
    USER_TYPING = "dm:user_typing",
    USER_STOPPED_TYPING = "dm:user_stopped_typing",
    READ_RECEIPT = "dm:read_receipt",
    USER_ONLINE = "dm:user_online",
    USER_OFFLINE = "dm:user_offline",
    CONVERSATION_UPDATED = "dm:conversation_updated",
}

export type ConversationParticipant = {
    userId: string;
    username: string;
    displayName: string | null;
    profilePicture: string | null;
    role: ParticipantRole;
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
    type: ConversationType;
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
    type: DmMessageType;
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
    type: DmClientEventType.SEND_MESSAGE;
    conversationId: string;
    text: string;
    messageType: DmMessageType.TEXT | DmMessageType.IMAGE;
    senderUsername: string;
    imageUrls?: string[];
    replyTo?: string;
};

export type DmWsTyping = {
    type: DmClientEventType.TYPING_START | DmClientEventType.TYPING_STOP;
    conversationId: string;
    username: string;
};

export type DmWsMarkRead = {
    type: DmClientEventType.MARK_READ;
    conversationId: string;
    messageId: string;
};

export type DmWsClientEvent = DmWsSendMessage | DmWsTyping | DmWsMarkRead;

// WebSocket event types (server → client)
export type DmWsNewMessage = {
    type: DmServerEventType.NEW_MESSAGE;
    conversationId: string;
    message: DmMessage;
};

export type DmWsMessageEdited = {
    type: DmServerEventType.MESSAGE_EDITED;
    conversationId: string;
    messageId: string;
    newText: string;
    updatedAt: string;
};

export type DmWsMessageDeleted = {
    type: DmServerEventType.MESSAGE_DELETED;
    conversationId: string;
    messageId: string;
};

export type DmWsUserTyping = {
    type: DmServerEventType.USER_TYPING | DmServerEventType.USER_STOPPED_TYPING;
    conversationId: string;
    userId: string;
    username: string;
};

export type DmWsReadReceipt = {
    type: DmServerEventType.READ_RECEIPT;
    conversationId: string;
    userId: string;
    messageId: string;
    readAt: string;
};

export type DmWsPresence = {
    type: DmServerEventType.USER_ONLINE | DmServerEventType.USER_OFFLINE;
    userId: string;
};

export type DmWsConversationUpdated = {
    type: DmServerEventType.CONVERSATION_UPDATED;
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
