export type ChatMessage = {
    type: ChatMessageType // type of the message "join", "leave", "message"
    roomId: string // room id
    userId: string
    username: string
    text: string // message text
}

export enum ChatMessageType {
    JOIN = 'join',
    LEAVE = 'leave',
    MESSAGE = 'message'
}
