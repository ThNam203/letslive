import { DmRedisService } from './dmRedis'

export class PresenceService {
    private redis: import('ioredis').default

    constructor(redis: import('ioredis').default) {
        this.redis = redis
    }

    async setOnline(userId: string) {
        await this.redis.sadd('dm:online_users', userId)
        await this.redis.set(`dm:user:${userId}:last_seen`, Date.now().toString())
    }

    async setOffline(userId: string) {
        await this.redis.srem('dm:online_users', userId)
        await this.redis.set(`dm:user:${userId}:last_seen`, Date.now().toString())
    }

    async isOnline(userId: string): Promise<boolean> {
        return (await this.redis.sismember('dm:online_users', userId)) === 1
    }

    async getOnlineUsers(userIds: string[]): Promise<string[]> {
        if (userIds.length === 0) return []
        const pipeline = this.redis.pipeline()
        for (const id of userIds) {
            pipeline.sismember('dm:online_users', id)
        }
        const results = await pipeline.exec()
        if (!results) return []

        return userIds.filter((_, i) => results[i] && results[i][1] === 1)
    }

    async getLastSeen(userId: string): Promise<number | null> {
        const val = await this.redis.get(`dm:user:${userId}:last_seen`)
        return val ? parseInt(val) : null
    }
}
