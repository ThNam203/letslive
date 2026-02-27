import mongoose, { Types } from 'mongoose'

export interface IParticipant {
    userId: string
    username: string
    displayName: string | null
    profilePicture: string | null
    role: 'owner' | 'admin' | 'member'
    joinedAt: Date
    lastReadMessageId: Types.ObjectId | null
    isMuted: boolean
}

export interface ILastMessage {
    _id: Types.ObjectId
    senderId: string
    senderUsername: string
    text: string
    createdAt: Date
}

export interface IConversation {
    _id: Types.ObjectId
    type: 'dm' | 'group'
    name: string | null
    avatarUrl: string | null
    createdBy: string
    participants: IParticipant[]
    lastMessage: ILastMessage | null
    createdAt: Date
    updatedAt: Date
}

const participantSchema = new mongoose.Schema(
    {
        userId: { type: String, required: true, maxlength: 36 },
        username: { type: String, required: true, maxlength: 50 },
        displayName: { type: String, default: null, maxlength: 50 },
        profilePicture: { type: String, default: null, maxlength: 2048 },
        role: { type: String, required: true, enum: ['owner', 'admin', 'member'], default: 'member' },
        joinedAt: { type: Date, default: Date.now },
        lastReadMessageId: { type: mongoose.Schema.Types.ObjectId, default: null },
        isMuted: { type: Boolean, default: false }
    },
    { _id: false }
)

const lastMessageSchema = new mongoose.Schema(
    {
        _id: { type: mongoose.Schema.Types.ObjectId, required: true },
        senderId: { type: String, required: true, maxlength: 36 },
        senderUsername: { type: String, required: true, maxlength: 50 },
        text: { type: String, required: true, maxlength: 100 },
        createdAt: { type: Date, required: true }
    },
    { _id: false }
)

const conversationSchema = new mongoose.Schema(
    {
        type: { type: String, required: true, enum: ['dm', 'group'] },
        name: { type: String, default: null, maxlength: 100 },
        avatarUrl: { type: String, default: null, maxlength: 2048 },
        createdBy: { type: String, required: true, maxlength: 36 },
        participants: { type: [participantSchema], required: true },
        lastMessage: { type: lastMessageSchema, default: null }
    },
    { timestamps: true }
)

conversationSchema.index({ 'participants.userId': 1, updatedAt: -1 })
conversationSchema.index({ type: 1, 'participants.userId': 1 })

const Conversation = mongoose.model<IConversation>('Conversation', conversationSchema)

export { Conversation }
