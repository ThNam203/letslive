import { Types } from 'mongoose'
import { DmMessage, IDmMessage } from '../models/DmMessage'
import { Conversation } from '../models/Conversation'
import { RESPONSE_TEMPLATES, Response as ServiceResponse, newResponseFromTemplate } from '../types/api-response'
import { DmMessageType } from '../types/conversation'
import logger from 'lib/logger'

export class DmMessageService {
    async sendMessage(
        conversationId: string,
        senderId: string,
        senderUsername: string,
        text: string,
        type: DmMessageType = DmMessageType.TEXT,
        imageUrls?: string[],
        replyTo?: string
    ): Promise<ServiceResponse<IDmMessage & { participantIds: string[] }>> {
        if (!Types.ObjectId.isValid(conversationId)) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_CONVERSATION_NOT_FOUND)
        }

        const conversation = await Conversation.findById(conversationId)
        if (!conversation) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_CONVERSATION_NOT_FOUND)
        }

        const isParticipant = conversation.participants.some((p) => p.userId === senderId)
        if (!isParticipant) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_NOT_PARTICIPANT)
        }

        if (!text || text.trim().length === 0) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT)
        }

        if (text.length > 2000) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT)
        }

        const message = new DmMessage({
            conversationId: new Types.ObjectId(conversationId),
            senderId,
            senderUsername,
            type,
            text: text.trim(),
            imageUrls: type === DmMessageType.IMAGE && imageUrls ? imageUrls : [],
            replyTo: replyTo && Types.ObjectId.isValid(replyTo) ? new Types.ObjectId(replyTo) : null,
            isDeleted: false,
            readBy: [{ userId: senderId, readAt: new Date() }]
        })

        await message.save()

        // Update conversation's lastMessage and updatedAt
        conversation.lastMessage = {
            _id: message._id,
            senderId,
            senderUsername,
            text: text.trim().substring(0, 100),
            createdAt: message.createdAt
        }
        await conversation.save()

        const participantIds = conversation.participants.map((p) => p.userId)

        return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_SUCC_CREATED, {
            ...message.toObject(),
            participantIds
        })
    }

    async getMessages(
        conversationId: string,
        userId: string,
        before?: string,
        limit: number = 50
    ): Promise<ServiceResponse<IDmMessage[]>> {
        if (!Types.ObjectId.isValid(conversationId)) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_CONVERSATION_NOT_FOUND)
        }

        const conversation = await Conversation.findById(conversationId)
        if (!conversation) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_CONVERSATION_NOT_FOUND)
        }

        const isParticipant = conversation.participants.some((p) => p.userId === userId)
        if (!isParticipant) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_NOT_PARTICIPANT)
        }

        const clampedLimit = Math.min(Math.max(limit, 1), 100)

        const query: Record<string, any> = {
            conversationId: new Types.ObjectId(conversationId)
        }

        if (before && Types.ObjectId.isValid(before)) {
            query._id = { $lt: new Types.ObjectId(before) }
        }

        const messages = await DmMessage.find(query).sort({ createdAt: -1 }).limit(clampedLimit)

        // Return in chronological order
        messages.reverse()

        return newResponseFromTemplate<IDmMessage[]>(RESPONSE_TEMPLATES.RES_SUCC_OK, messages.map((m) => m.toObject()))
    }

    async deleteMessage(
        conversationId: string,
        messageId: string,
        userId: string
    ): Promise<ServiceResponse<void>> {
        if (!Types.ObjectId.isValid(conversationId) || !Types.ObjectId.isValid(messageId)) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_DM_MESSAGE_NOT_FOUND)
        }

        const message = await DmMessage.findOne({
            _id: new Types.ObjectId(messageId),
            conversationId: new Types.ObjectId(conversationId)
        })

        if (!message) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_DM_MESSAGE_NOT_FOUND)
        }

        if (message.senderId !== userId) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_FORBIDDEN)
        }

        message.isDeleted = true
        message.text = ''
        await message.save()

        return newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_SUCC_OK)
    }

    async editMessage(
        conversationId: string,
        messageId: string,
        userId: string,
        newText: string
    ): Promise<ServiceResponse<IDmMessage>> {
        if (!Types.ObjectId.isValid(conversationId) || !Types.ObjectId.isValid(messageId)) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_DM_MESSAGE_NOT_FOUND)
        }

        if (!newText || newText.trim().length === 0 || newText.length > 2000) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT)
        }

        const message = await DmMessage.findOne({
            _id: new Types.ObjectId(messageId),
            conversationId: new Types.ObjectId(conversationId)
        })

        if (!message) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_DM_MESSAGE_NOT_FOUND)
        }

        if (message.senderId !== userId) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_FORBIDDEN)
        }

        if (message.isDeleted) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_DM_MESSAGE_NOT_FOUND)
        }

        message.text = newText.trim()
        await message.save()

        return newResponseFromTemplate<IDmMessage>(RESPONSE_TEMPLATES.RES_SUCC_OK, message.toObject())
    }

    async markAsRead(
        conversationId: string,
        userId: string,
        messageId?: string
    ): Promise<ServiceResponse<void>> {
        if (!Types.ObjectId.isValid(conversationId)) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_CONVERSATION_NOT_FOUND)
        }

        const conversation = await Conversation.findById(conversationId)
        if (!conversation) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_CONVERSATION_NOT_FOUND)
        }

        const participant = conversation.participants.find((p) => p.userId === userId)
        if (!participant) {
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_NOT_PARTICIPANT)
        }

        // Find the latest message if no specific messageId provided
        let readUpToId: Types.ObjectId
        if (messageId && Types.ObjectId.isValid(messageId)) {
            readUpToId = new Types.ObjectId(messageId)
        } else {
            const latestMessage = await DmMessage.findOne({
                conversationId: new Types.ObjectId(conversationId)
            }).sort({ createdAt: -1 })
            if (!latestMessage) {
                return newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_SUCC_OK)
            }
            readUpToId = latestMessage._id
        }

        participant.lastReadMessageId = readUpToId
        await conversation.save()

        return newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_SUCC_OK)
    }
}
