import { Types } from 'mongoose'
import { Conversation, IConversation } from '../models/Conversation'
import { DmMessage } from '../models/DmMessage'
import { RESPONSE_TEMPLATES, Response as ServiceResponse, newResponseFromTemplate } from '../types/api-response'
import logger from 'lib/logger'

const MAX_GROUP_PARTICIPANTS = 50

export class ConversationService {
    async createConversation(
        type: 'dm' | 'group',
        creatorId: string,
        creatorUsername: string,
        creatorDisplayName: string | null,
        creatorProfilePicture: string | null,
        participantInfos: Array<{
            userId: string
            username: string
            displayName: string | null
            profilePicture: string | null
        }>,
        name?: string
    ): Promise<ServiceResponse<IConversation>> {
        if (type === 'dm') {
            if (participantInfos.length !== 1) {
                return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT)
            }

            if (participantInfos[0].userId === creatorId) {
                return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_CANNOT_MESSAGE_SELF)
            }

            // Check if DM already exists between these two users
            const existing = await Conversation.findOne({
                type: 'dm',
                'participants.userId': { $all: [creatorId, participantInfos[0].userId] },
                $expr: { $eq: [{ $size: '$participants' }, 2] }
            })

            if (existing) {
                return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_SUCC_OK, existing.toObject())
            }
        }

        if (type === 'group') {
            if (participantInfos.length < 1) {
                return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT)
            }
            if (participantInfos.length + 1 > MAX_GROUP_PARTICIPANTS) {
                return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_TOO_MANY_PARTICIPANTS)
            }
        }

        const participants = [
            {
                userId: creatorId,
                username: creatorUsername,
                displayName: creatorDisplayName,
                profilePicture: creatorProfilePicture,
                role: type === 'group' ? ('owner' as const) : ('member' as const),
                joinedAt: new Date(),
                lastReadMessageId: null,
                isMuted: false
            },
            ...participantInfos.map((p) => ({
                userId: p.userId,
                username: p.username,
                displayName: p.displayName,
                profilePicture: p.profilePicture,
                role: 'member' as const,
                joinedAt: new Date(),
                lastReadMessageId: null,
                isMuted: false
            }))
        ]

        const conversation = new Conversation({
            type,
            name: type === 'group' ? (name || null) : null,
            avatarUrl: null,
            createdBy: creatorId,
            participants,
            lastMessage: null
        })

        await conversation.save()
        return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_SUCC_CREATED, conversation.toObject())
    }

    async getConversations(userId: string, page: number, limit: number): Promise<ServiceResponse<IConversation[]>> {
        const skip = page * limit
        const conversations = await Conversation.find({ 'participants.userId': userId })
            .sort({ updatedAt: -1 })
            .skip(skip)
            .limit(limit)

        const total = await Conversation.countDocuments({ 'participants.userId': userId })

        return newResponseFromTemplate<IConversation[]>(
            RESPONSE_TEMPLATES.RES_SUCC_OK,
            conversations.map((c) => c.toObject()),
            { page, page_size: limit, total }
        )
    }

    async getConversation(conversationId: string, userId: string): Promise<ServiceResponse<IConversation>> {
        if (!Types.ObjectId.isValid(conversationId)) {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_CONVERSATION_NOT_FOUND)
        }

        const conversation = await Conversation.findById(conversationId)
        if (!conversation) {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_CONVERSATION_NOT_FOUND)
        }

        const isParticipant = conversation.participants.some((p) => p.userId === userId)
        if (!isParticipant) {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_NOT_PARTICIPANT)
        }

        return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_SUCC_OK, conversation.toObject())
    }

    async updateConversation(
        conversationId: string,
        userId: string,
        updates: { name?: string; avatarUrl?: string }
    ): Promise<ServiceResponse<IConversation>> {
        if (!Types.ObjectId.isValid(conversationId)) {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_CONVERSATION_NOT_FOUND)
        }

        const conversation = await Conversation.findById(conversationId)
        if (!conversation) {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_CONVERSATION_NOT_FOUND)
        }

        if (conversation.type !== 'group') {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_FORBIDDEN)
        }

        const participant = conversation.participants.find((p) => p.userId === userId)
        if (!participant) {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_NOT_PARTICIPANT)
        }

        if (participant.role !== 'owner' && participant.role !== 'admin') {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_INSUFFICIENT_ROLE)
        }

        if (updates.name !== undefined) conversation.name = updates.name
        if (updates.avatarUrl !== undefined) conversation.avatarUrl = updates.avatarUrl
        await conversation.save()

        return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_SUCC_OK, conversation.toObject())
    }

    async addParticipant(
        conversationId: string,
        userId: string,
        newParticipant: {
            userId: string
            username: string
            displayName: string | null
            profilePicture: string | null
        }
    ): Promise<ServiceResponse<IConversation>> {
        if (!Types.ObjectId.isValid(conversationId)) {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_CONVERSATION_NOT_FOUND)
        }

        const conversation = await Conversation.findById(conversationId)
        if (!conversation) {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_CONVERSATION_NOT_FOUND)
        }

        if (conversation.type !== 'group') {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_FORBIDDEN)
        }

        const actor = conversation.participants.find((p) => p.userId === userId)
        if (!actor) {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_NOT_PARTICIPANT)
        }

        if (actor.role !== 'owner' && actor.role !== 'admin') {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_INSUFFICIENT_ROLE)
        }

        const alreadyParticipant = conversation.participants.some((p) => p.userId === newParticipant.userId)
        if (alreadyParticipant) {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_SUCC_OK, conversation.toObject())
        }

        if (conversation.participants.length >= MAX_GROUP_PARTICIPANTS) {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_TOO_MANY_PARTICIPANTS)
        }

        conversation.participants.push({
            userId: newParticipant.userId,
            username: newParticipant.username,
            displayName: newParticipant.displayName,
            profilePicture: newParticipant.profilePicture,
            role: 'member',
            joinedAt: new Date(),
            lastReadMessageId: null,
            isMuted: false
        })

        await conversation.save()
        return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_SUCC_OK, conversation.toObject())
    }

    async removeParticipant(
        conversationId: string,
        userId: string,
        targetUserId: string
    ): Promise<ServiceResponse<IConversation>> {
        if (!Types.ObjectId.isValid(conversationId)) {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_CONVERSATION_NOT_FOUND)
        }

        const conversation = await Conversation.findById(conversationId)
        if (!conversation) {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_CONVERSATION_NOT_FOUND)
        }

        if (conversation.type !== 'group') {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_FORBIDDEN)
        }

        const actor = conversation.participants.find((p) => p.userId === userId)
        if (!actor) {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_NOT_PARTICIPANT)
        }

        if (actor.role !== 'owner' && actor.role !== 'admin') {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_INSUFFICIENT_ROLE)
        }

        const targetIdx = conversation.participants.findIndex((p) => p.userId === targetUserId)
        if (targetIdx === -1) {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_NOT_PARTICIPANT)
        }

        // Cannot remove the owner
        if (conversation.participants[targetIdx].role === 'owner') {
            return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_ERR_INSUFFICIENT_ROLE)
        }

        conversation.participants.splice(targetIdx, 1)
        await conversation.save()

        return newResponseFromTemplate<IConversation>(RESPONSE_TEMPLATES.RES_SUCC_OK, conversation.toObject())
    }

    async leaveConversation(conversationId: string, userId: string): Promise<ServiceResponse<void>> {
        if (!Types.ObjectId.isValid(conversationId)) {
            return newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_CONVERSATION_NOT_FOUND)
        }

        const conversation = await Conversation.findById(conversationId)
        if (!conversation) {
            return newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_CONVERSATION_NOT_FOUND)
        }

        const participantIdx = conversation.participants.findIndex((p) => p.userId === userId)
        if (participantIdx === -1) {
            return newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_NOT_PARTICIPANT)
        }

        if (conversation.type === 'dm') {
            // For DM, delete the conversation entirely
            await Conversation.deleteOne({ _id: conversation._id })
            await DmMessage.deleteMany({ conversationId: conversation._id })
        } else {
            // For group, remove the participant
            conversation.participants.splice(participantIdx, 1)

            if (conversation.participants.length === 0) {
                // No participants left, delete
                await Conversation.deleteOne({ _id: conversation._id })
                await DmMessage.deleteMany({ conversationId: conversation._id })
            } else {
                // If the leaving user was owner, transfer to the next admin or oldest member
                if (conversation.participants.every((p) => p.role !== 'owner')) {
                    const newOwner =
                        conversation.participants.find((p) => p.role === 'admin') || conversation.participants[0]
                    newOwner.role = 'owner'
                }
                await conversation.save()
            }
        }

        return newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_SUCC_OK)
    }

    async getUnreadCounts(userId: string): Promise<ServiceResponse<Record<string, number>>> {
        const conversations = await Conversation.find({ 'participants.userId': userId })
        const unreadCounts: Record<string, number> = {}

        for (const conv of conversations) {
            const participant = conv.participants.find((p) => p.userId === userId)
            if (!participant) continue

            if (participant.lastReadMessageId) {
                const count = await DmMessage.countDocuments({
                    conversationId: conv._id,
                    _id: { $gt: participant.lastReadMessageId },
                    senderId: { $ne: userId },
                    isDeleted: false
                })
                if (count > 0) {
                    unreadCounts[conv._id.toString()] = count
                }
            } else {
                // Never read — count all messages not from this user
                const count = await DmMessage.countDocuments({
                    conversationId: conv._id,
                    senderId: { $ne: userId },
                    isDeleted: false
                })
                if (count > 0) {
                    unreadCounts[conv._id.toString()] = count
                }
            }
        }

        return newResponseFromTemplate<Record<string, number>>(RESPONSE_TEMPLATES.RES_SUCC_OK, unreadCounts)
    }

    async isParticipant(conversationId: string, userId: string): Promise<boolean> {
        if (!Types.ObjectId.isValid(conversationId)) return false
        const conversation = await Conversation.findById(conversationId)
        if (!conversation) return false
        return conversation.participants.some((p) => p.userId === userId)
    }

    async getParticipantIds(conversationId: string): Promise<string[]> {
        if (!Types.ObjectId.isValid(conversationId)) return []
        const conversation = await Conversation.findById(conversationId)
        if (!conversation) return []
        return conversation.participants.map((p) => p.userId)
    }
}
