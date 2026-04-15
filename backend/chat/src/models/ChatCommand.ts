import mongoose, { Types } from 'mongoose'

export const CHAT_COMMAND_SCOPE_USER = 'user'
export const CHAT_COMMAND_SCOPE_CHANNEL = 'channel'

export type ChatCommandScope = typeof CHAT_COMMAND_SCOPE_USER | typeof CHAT_COMMAND_SCOPE_CHANNEL

export interface IChatCommand {
    _id: Types.ObjectId
    scope: ChatCommandScope
    ownerId: string
    name: string
    response: string
    description: string
    createdAt: Date
}

const chatCommandSchema = new mongoose.Schema(
    {
        scope: {
            type: String,
            required: true,
            enum: [CHAT_COMMAND_SCOPE_USER, CHAT_COMMAND_SCOPE_CHANNEL],
            index: true
        },
        ownerId: {
            type: String,
            required: true,
            maxlength: 36,
            index: true
        },
        name: {
            type: String,
            required: true,
            maxlength: 32,
            lowercase: true,
            trim: true,
            match: /^[a-z0-9_-]+$/
        },
        response: {
            type: String,
            required: true,
            maxlength: 500
        },
        description: {
            type: String,
            maxlength: 120,
            default: ''
        },
        createdAt: { type: Date, default: Date.now }
    },
    { collection: 'chat_commands' }
)

chatCommandSchema.index({ scope: 1, ownerId: 1, name: 1 }, { unique: true })

const ChatCommand = mongoose.model<IChatCommand>('ChatCommand', chatCommandSchema)

export { ChatCommand }
