export type ChatMessage = {
    type: ChatMessageType // type of the message "join", "leave", "message"
    room: string // room id
    senderName: string
    senderId: string
    text: string // message text
}

export enum ChatMessageType {
    JOIN = 'join',
    LEAVE = 'leave',
    MESSAGE = 'message'
}
