export type CreateConversationRequest = {
    type: 'dm' | 'group'
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
    type: 'text' | 'image'
    imageUrls?: string[]
    replyTo?: string
}

export type EditDmMessageRequest = {
    text: string
}
