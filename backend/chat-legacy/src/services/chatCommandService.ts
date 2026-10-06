import {
    ChatCommand,
    CHAT_COMMAND_SCOPE_USER,
    CHAT_COMMAND_SCOPE_CHANNEL,
    ChatCommandScope,
    IChatCommand
} from '../models/ChatCommand'
import { RESPONSE_TEMPLATES, Response as ServiceResponse, newResponseFromTemplate } from '../types/api-response'
import logger from 'lib/logger'

export type { ChatCommandScope }

const NAME_PATTERN = /^[a-z0-9_-]{1,32}$/
const MAX_RESPONSE = 500
const MAX_DESCRIPTION = 120
const MAX_PER_OWNER_PER_SCOPE = 50

export interface ChatCommandDoc {
    id: string
    scope: ChatCommandScope
    ownerId: string
    name: string
    response: string
    description: string
    createdAt: Date
}

function toDoc(d: IChatCommand): ChatCommandDoc {
    return {
        id: d._id.toString(),
        scope: d.scope,
        ownerId: d.ownerId,
        name: d.name,
        response: d.response,
        description: d.description ?? '',
        createdAt: d.createdAt
    }
}

export class ChatCommandService {
    async listForRoom(roomId: string, userId: string | null): Promise<ServiceResponse<ChatCommandDoc[]>> {
        try {
            const filters: any[] = [{ scope: CHAT_COMMAND_SCOPE_CHANNEL, ownerId: roomId }]
            if (userId) filters.push({ scope: CHAT_COMMAND_SCOPE_USER, ownerId: userId })

            const docs = await ChatCommand.find({ $or: filters }).sort({ name: 1 })
            return newResponseFromTemplate<ChatCommandDoc[]>(RESPONSE_TEMPLATES.RES_SUCC_OK, docs.map(toDoc))
        } catch (err) {
            logger.error({ err }, 'failed to list chat commands for room')
            return newResponseFromTemplate<ChatCommandDoc[]>(RESPONSE_TEMPLATES.RES_ERR_DATABASE_QUERY)
        }
    }

    async listMine(
        userId: string
    ): Promise<ServiceResponse<{ user: ChatCommandDoc[]; channel: ChatCommandDoc[] }>> {
        try {
            const docs = await ChatCommand.find({
                $or: [
                    { scope: CHAT_COMMAND_SCOPE_USER, ownerId: userId },
                    { scope: CHAT_COMMAND_SCOPE_CHANNEL, ownerId: userId }
                ]
            }).sort({ name: 1 })

            const user: ChatCommandDoc[] = []
            const channel: ChatCommandDoc[] = []
            for (const d of docs) {
                const doc = toDoc(d)
                if (doc.scope === CHAT_COMMAND_SCOPE_USER) user.push(doc)
                else channel.push(doc)
            }
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_SUCC_OK, { user, channel })
        } catch (err) {
            logger.error({ err }, 'failed to list own chat commands')
            return newResponseFromTemplate(RESPONSE_TEMPLATES.RES_ERR_DATABASE_QUERY)
        }
    }

    async create(
        userId: string,
        scope: ChatCommandScope,
        name: string,
        response: string,
        description: string
    ): Promise<ServiceResponse<ChatCommandDoc>> {
        const normalized = (name || '').trim().toLowerCase()
        if (!NAME_PATTERN.test(normalized)) {
            return newResponseFromTemplate<ChatCommandDoc>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT)
        }
        if (typeof response !== 'string' || response.length === 0 || response.length > MAX_RESPONSE) {
            return newResponseFromTemplate<ChatCommandDoc>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT)
        }
        if (typeof description !== 'string' || description.length > MAX_DESCRIPTION) {
            return newResponseFromTemplate<ChatCommandDoc>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT)
        }
        if (scope !== CHAT_COMMAND_SCOPE_USER && scope !== CHAT_COMMAND_SCOPE_CHANNEL) {
            return newResponseFromTemplate<ChatCommandDoc>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT)
        }

        try {
            const count = await ChatCommand.countDocuments({ scope, ownerId: userId })
            if (count >= MAX_PER_OWNER_PER_SCOPE) {
                return newResponseFromTemplate<ChatCommandDoc>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT)
            }

            const created = await ChatCommand.create({
                scope,
                ownerId: userId,
                name: normalized,
                response,
                description
            })
            return newResponseFromTemplate<ChatCommandDoc>(RESPONSE_TEMPLATES.RES_SUCC_CREATED, toDoc(created))
        } catch (err: any) {
            if (err && err.code === 11000) {
                return newResponseFromTemplate<ChatCommandDoc>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT)
            }
            logger.error({ err }, 'failed to create chat command')
            return newResponseFromTemplate<ChatCommandDoc>(RESPONSE_TEMPLATES.RES_ERR_DATABASE_QUERY)
        }
    }

    async update(
        userId: string,
        id: string,
        name: string | undefined,
        response: string | undefined,
        description: string | undefined
    ): Promise<ServiceResponse<ChatCommandDoc>> {
        try {
            const doc = await ChatCommand.findById(id)
            if (!doc) {
                return newResponseFromTemplate<ChatCommandDoc>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT)
            }
            if (doc.ownerId !== userId) {
                return newResponseFromTemplate<ChatCommandDoc>(RESPONSE_TEMPLATES.RES_ERR_FORBIDDEN)
            }

            if (name !== undefined) {
                const normalized = (name || '').trim().toLowerCase()
                if (!NAME_PATTERN.test(normalized)) {
                    return newResponseFromTemplate<ChatCommandDoc>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT)
                }
                doc.name = normalized
            }
            if (response !== undefined) {
                if (typeof response !== 'string' || response.length === 0 || response.length > MAX_RESPONSE) {
                    return newResponseFromTemplate<ChatCommandDoc>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT)
                }
                doc.response = response
            }
            if (description !== undefined) {
                if (typeof description !== 'string' || description.length > MAX_DESCRIPTION) {
                    return newResponseFromTemplate<ChatCommandDoc>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT)
                }
                doc.description = description
            }

            await doc.save()
            return newResponseFromTemplate<ChatCommandDoc>(RESPONSE_TEMPLATES.RES_SUCC_OK, toDoc(doc))
        } catch (err: any) {
            if (err && err.code === 11000) {
                return newResponseFromTemplate<ChatCommandDoc>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT)
            }
            logger.error({ err }, 'failed to update chat command')
            return newResponseFromTemplate<ChatCommandDoc>(RESPONSE_TEMPLATES.RES_ERR_DATABASE_QUERY)
        }
    }

    async delete(userId: string, id: string): Promise<ServiceResponse<void>> {
        try {
            const doc = await ChatCommand.findById(id)
            if (!doc) return newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_SUCC_OK)
            if (doc.ownerId !== userId) {
                return newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_FORBIDDEN)
            }
            await doc.deleteOne()
            return newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_SUCC_OK)
        } catch (err) {
            logger.error({ err }, 'failed to delete chat command')
            return newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_DATABASE_QUERY)
        }
    }
}
