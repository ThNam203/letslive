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
    imageUrls: string[];
    replyTo: string | null;
    isDeleted: boolean;
    readBy: ReadReceipt[];
    createdAt: string;
    updatedAt: string;
};

// Realtime DM events (server → client). The gateway delivers `type` in the
// frame and the rest as its data.
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
    /** absent when the client marked read without naming a message */
    messageId?: string;
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
