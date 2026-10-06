import WebSocket, { WebSocketServer } from 'ws'
import { DmRedisService } from './services/dmRedis'
import { PresenceService } from './services/presenceService'
import { ConversationService } from './services/conversationService'
import { DmMessageService } from './services/dmMessageService'
import {
    DmClientEvent,
    DmClientEventType,
    DmMarkReadEvent,
    DmSendMessageEvent,
    DmServerEventType,
    DmTypingEvent
} from './types/dm-event'
import { DmMessageType } from './types/conversation'
import logger from './lib/logger'

export class DmServer {
    private connections: Map<string, WebSocket> = new Map()

    constructor(
        private dmRedisService: DmRedisService,
        private presenceService: PresenceService,
        private conversationService: ConversationService,
        private dmMessageService: DmMessageService,
        private wss: WebSocketServer
    ) {
        this.initialize()
    }

    private initialize() {
        this.wss.on('connection', (ws, req) => {
            const userId = (req as any).userId as string
            this.handleConnection(ws, userId)
        })

        this.dmRedisService.subscribeToMessages((pattern, channel, message) => {
            this.handleRedisMessage(channel, message)
        })
    }

    private async handleConnection(ws: WebSocket, userId: string) {
        // Store connection
        this.connections.set(userId, ws)
        await this.presenceService.setOnline(userId)

        // Broadcast online status to contacts
        this.broadcastPresence(userId, DmServerEventType.USER_ONLINE)

        ws.on('message', async (rawMessage) => {
            const raw = rawMessage.toString()
            if (raw.length > 8192) {
                logger.error('DM WebSocket message too large, dropping')
                return
            }

            try {
                const data = JSON.parse(raw)
                await this.handleClientEvent(data, userId)
            } catch (err) {
                logger.error({ err }, 'failed to parse DM WebSocket message')
            }
        })

        ws.on('close', async () => {
            this.connections.delete(userId)
            await this.presenceService.setOffline(userId)
            this.broadcastPresence(userId, DmServerEventType.USER_OFFLINE)
        })

        // Keep-alive ping every 30s
        const pingInterval = setInterval(() => {
            if (ws.readyState === WebSocket.OPEN) {
                ws.ping()
            } else {
                clearInterval(pingInterval)
            }
        }, 30000)
    }

    private async handleClientEvent(data: unknown, userId: string) {
        if (typeof data !== 'object' || data === null || !('type' in data)) {
            logger.warn('malformed DM event, dropping')
            return
        }

        const event = data as DmClientEvent
        switch (event.type) {
            case DmClientEventType.SEND_MESSAGE:
                await this.handleSendMessage(event, userId)
                break
            case DmClientEventType.TYPING_START:
            case DmClientEventType.TYPING_STOP:
                await this.handleTyping(event, userId)
                break
            case DmClientEventType.MARK_READ:
                await this.handleMarkRead(event, userId)
                break
            default:
                logger.warn({ type: (event as { type: string }).type }, 'unknown DM event type')
        }
    }

    private async handleSendMessage(data: DmSendMessageEvent, userId: string) {
        const { conversationId, text, messageType, imageUrls, replyTo } = data

        if (!conversationId || !text || typeof text !== 'string') {
            return
        }

        if (text.length > 2000) {
            return
        }

        const result = await this.dmMessageService.sendMessage(
            conversationId,
            userId,
            text,
            messageType || DmMessageType.TEXT,
            imageUrls,
            replyTo
        )

        if (!result.success || !result.data) {
            const ws = this.connections.get(userId)
            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.send(
                    JSON.stringify({
                        type: DmServerEventType.SEND_FAILED,
                        key: result.key ?? 'res_err_invalid_input',
                        message: result.message
                    })
                )
            }
            return
        }

        const { participantIds, ...message } = result.data

        // Publish to Redis for multi-instance support
        await this.dmRedisService.publishMessage(conversationId, {
            type: DmServerEventType.NEW_MESSAGE,
            conversationId,
            message,
            recipientIds: participantIds
        })
    }

    private async handleTyping(data: DmTypingEvent, userId: string) {
        const { conversationId, type } = data

        if (!conversationId) return

        const username = await this.conversationService.getParticipantUsername(conversationId, userId)
        if (username === null) return

        const participantIds = await this.conversationService.getParticipantIds(conversationId)

        const eventType =
            type === DmClientEventType.TYPING_START
                ? DmServerEventType.USER_TYPING
                : DmServerEventType.USER_STOPPED_TYPING

        // Publish typing event via Redis (ephemeral, not persisted)
        await this.dmRedisService.publishEvent(conversationId, {
            type: eventType,
            conversationId,
            userId,
            username,
            recipientIds: participantIds.filter((id) => id !== userId)
        })
    }

    private async handleMarkRead(data: DmMarkReadEvent, userId: string) {
        const { conversationId, messageId } = data

        if (!conversationId) return

        const result = await this.dmMessageService.markAsRead(conversationId, userId, messageId)
        if (!result.success) return

        const participantIds = await this.conversationService.getParticipantIds(conversationId)

        // Publish read receipt via Redis
        await this.dmRedisService.publishEvent(conversationId, {
            type: DmServerEventType.READ_RECEIPT,
            conversationId,
            userId,
            messageId,
            readAt: new Date().toISOString(),
            recipientIds: participantIds.filter((id) => id !== userId)
        })
    }

    private async handleRedisMessage(channel: string, message: string) {
        try {
            const data = JSON.parse(message)
            const recipientIds: string[] = data.recipientIds || []

            const payload = { ...data }
            delete payload.recipientIds

            for (const id of recipientIds) {
                const ws = this.connections.get(id)
                if (ws && ws.readyState === WebSocket.OPEN) {
                    ws.send(JSON.stringify(payload))
                }
            }
        } catch (err) {
            logger.error({ err }, 'failed to handle DM Redis message')
        }
    }

    private async broadcastPresence(userId: string, eventType: DmServerEventType) {
        // Get all conversations for this user to find contacts
        // This is a lightweight operation since we only need participant IDs
        const contactIds = new Set<string>()

        // Find all conversations this user is part of
        const { data: conversations } = await this.conversationService.getConversations(userId, 0, 100)

        if (conversations) {
            for (const conv of conversations) {
                for (const p of conv.participants) {
                    if (p.userId !== userId) {
                        contactIds.add(p.userId)
                    }
                }
            }
        }

        const payload = JSON.stringify({
            type: eventType,
            userId
        })

        for (const contactId of contactIds) {
            const ws = this.connections.get(contactId)
            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.send(payload)
            }
        }
    }
}
