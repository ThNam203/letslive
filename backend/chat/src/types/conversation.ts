export enum ConversationType {
    DM = 'dm',
    GROUP = 'group'
}

export enum DmMessageType {
    TEXT = 'text',
    IMAGE = 'image',
    SYSTEM = 'system'
}

export enum ParticipantRole {
    OWNER = 'owner',
    ADMIN = 'admin',
    MEMBER = 'member'
}

export type CreateConversationRequest = {
    type: ConversationType
    participantIds: string[]
    name?: string
}

export type UpdateConversationRequest = {
    name?: string
    avatarUrl?: string
}

export type AddParticipantRequest = {
    userId: string
    username: string
    displayName?: string
    profilePicture?: string
}

export type SendDmMessageRequest = {
    text: string
    type: DmMessageType
    imageUrls?: string[]
    replyTo?: string
}

export type EditDmMessageRequest = {
    text: string
}
