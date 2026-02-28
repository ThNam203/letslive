import Redis from 'ioredis'
import logger from '../lib/logger'

export class DmRedisService {
    constructor(
        private pub: Redis,
        private sub: Redis,
        private manager: Redis
    ) {}

    async publishMessage(conversationId: string, data: Record<string, any>) {
        await this.pub.publish(`dm:${conversationId}:messages`, JSON.stringify(data))
    }

    async publishEvent(conversationId: string, data: Record<string, any>) {
        await this.pub.publish(`dm:${conversationId}:events`, JSON.stringify(data))
    }

    subscribeToMessages(callback: (pattern: string, channel: string, message: string) => void) {
        this.sub.on('pmessage', callback)
        this.sub.psubscribe('dm:*:messages', 'dm:*:events')
    }

    // Online presence
    async setOnline(userId: string) {
        await this.manager.sadd('dm:online_users', userId)
        await this.manager.set(`dm:user:${userId}:last_seen`, Date.now().toString())
    }

    async setOffline(userId: string) {
        await this.manager.srem('dm:online_users', userId)
        await this.manager.set(`dm:user:${userId}:last_seen`, Date.now().toString())
    }

    async isOnline(userId: string): Promise<boolean> {
        return (await this.manager.sismember('dm:online_users', userId)) === 1
    }

    async getOnlineUsers(userIds: string[]): Promise<string[]> {
        const pipeline = this.manager.pipeline()
        for (const id of userIds) {
            pipeline.sismember('dm:online_users', id)
        }
        const results = await pipeline.exec()
        if (!results) return []

        return userIds.filter((_, i) => results[i] && results[i][1] === 1)
    }

    async getLastSeen(userId: string): Promise<number | null> {
        const val = await this.manager.get(`dm:user:${userId}:last_seen`)
        return val ? parseInt(val) : null
    }
}
