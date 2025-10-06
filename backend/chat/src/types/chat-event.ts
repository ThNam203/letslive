export type ChatEvent = {
    type: ChatEventType
    username: string | null
    userId: string | null
}

export enum ChatEventType {
    JOIN = 'join',
    LEAVE = 'leave'
}