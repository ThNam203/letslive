import mongoose, { Types } from 'mongoose'

export interface IReadReceipt {
    userId: string
    readAt: Date
}

export interface IDmMessage {
    _id: Types.ObjectId
    conversationId: Types.ObjectId
    senderId: string
    senderUsername: string
    type: 'text' | 'image' | 'system'
    text: string
    imageUrls: string[]
    replyTo: Types.ObjectId | null
    isDeleted: boolean
    readBy: IReadReceipt[]
    createdAt: Date
    updatedAt: Date
}

const readReceiptSchema = new mongoose.Schema(
    {
        userId: { type: String, required: true, maxlength: 36 },
        readAt: { type: Date, required: true, default: Date.now }
    },
    { _id: false }
)

const dmMessageSchema = new mongoose.Schema(
    {
        conversationId: { type: mongoose.Schema.Types.ObjectId, required: true, index: true },
        senderId: { type: String, required: true, maxlength: 36 },
        senderUsername: { type: String, required: true, maxlength: 50 },
        type: { type: String, required: true, enum: ['text', 'image', 'system'], default: 'text' },
        text: { type: String, required: true, maxlength: 2000 },
        imageUrls: { type: [String], default: [] },
        replyTo: { type: mongoose.Schema.Types.ObjectId, default: null },
        isDeleted: { type: Boolean, default: false },
        readBy: { type: [readReceiptSchema], default: [] }
    },
    { timestamps: true }
)

dmMessageSchema.index({ conversationId: 1, createdAt: -1 })

const DmMessage = mongoose.model<IDmMessage>('DmMessage', dmMessageSchema)

export { DmMessage }
