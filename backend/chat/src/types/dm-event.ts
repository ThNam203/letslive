// Client → Server events
export enum DmClientEventType {
    SEND_MESSAGE = 'dm:send_message',
    TYPING_START = 'dm:typing_start',
    TYPING_STOP = 'dm:typing_stop',
    MARK_READ = 'dm:mark_read'
}

export type DmSendMessageEvent = {
    type: DmClientEventType.SEND_MESSAGE
    conversationId: string
    text: string
    messageType: 'text' | 'image'
    imageUrls?: string[]
    replyTo?: string
}

export type DmTypingEvent = {
    type: DmClientEventType.TYPING_START | DmClientEventType.TYPING_STOP
    conversationId: string
}

export type DmMarkReadEvent = {
    type: DmClientEventType.MARK_READ
    conversationId: string
    messageId: string
}

export type DmClientEvent = DmSendMessageEvent | DmTypingEvent | DmMarkReadEvent

// Server → Client events
export enum DmServerEventType {
    NEW_MESSAGE = 'dm:new_message',
    MESSAGE_EDITED = 'dm:message_edited',
    MESSAGE_DELETED = 'dm:message_deleted',
    USER_TYPING = 'dm:user_typing',
    USER_STOPPED_TYPING = 'dm:user_stopped_typing',
    READ_RECEIPT = 'dm:read_receipt',
    USER_ONLINE = 'dm:user_online',
    USER_OFFLINE = 'dm:user_offline',
    CONVERSATION_UPDATED = 'dm:conversation_updated'
}
