export type ChatEvent = {
    type: ChatEventType
    username: string | null
    userId: string | null
    profilePicture: string | null
    timestamp: number
}

export enum ChatEventType {
    JOIN = 'join',
    LEAVE = 'leave'
}
