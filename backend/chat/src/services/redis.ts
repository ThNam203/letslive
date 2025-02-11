import Redis from 'ioredis'
import { ChatEvent, ChatEventType } from '~/types/chat_event'
import { ChatMessage } from '~/types/chat_message'

// src/services/RedisService.ts
export class RedisService {
    constructor(
        private pub: Redis,
        private sub: Redis,
        private roomManager: Redis
    ) {}

    async checkIfUserInRoom(userId: string, room: string) {
        return await this.roomManager.sismember(`room:${room}:members`, userId)
    }

    async addUserToRoom(userId: string, room: string) {
        await this.roomManager.sadd(`room:${room}:members`, userId)
    }

    async removeUserFromRoom(userId: string, room: string) {
        await this.roomManager.srem(`room:${room}:members`, userId)
    }

    async getUsersInRoom(room: string): Promise<string[]> {
        return await this.roomManager.smembers(`room:${room}:members`)
    }

    async publishEvent(room: string, event: ChatEventType, userId: string | undefined, userName: string | undefined) {
        const eventObj: ChatEvent = {
            type: event,
            userId: userId ?? null,
            username: userName ?? null
        }
        await this.pub.publish(`room:${room}:events`, JSON.stringify(eventObj))
    }

    async publishMessage(room: string, data: ChatMessage) {
        const timestamp = Date.now()
        await this.pub.publish(
            `room:${room}:messages`,
            JSON.stringify({
                senderId: data.senderId,
                senderName: data.senderName,
                text: data.text,
                timestamp
            })
        )
    }

    subscribeToMessages(callback: (pattern: string, channel: string, message: string) => void) {
        this.sub.on('pmessage', callback)
        this.sub.psubscribe('room:*:messages', 'room:*:events')
    }
}
