import ConsulRegistry from '../services/discovery'
import logger from '../lib/logger'

export type UserIdentity = {
    id: string
    username: string
    profilePicture: string | null
}

type BatchResponse = {
    data?: UserIdentity[]
}

const REQUEST_TIMEOUT_MS = 3000

/**
 * Resolves user identities through the user service.
 *
 * Display names must never be taken from the caller's payload: a client can put
 * any name there and it ends up persisted on conversations and messages. The
 * user service owns the username, so it is the only trusted source.
 */
export class UserServiceGateway {
    constructor(private registry: ConsulRegistry) {}

    async getIdentities(ids: string[]): Promise<Map<string, UserIdentity>> {
        const unique = [...new Set(ids)]
        if (unique.length === 0) return new Map()

        const address = await this.registry.serviceAddress('user')

        const controller = new AbortController()
        const timeout = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS)

        try {
            const res = await fetch(`http://${address}/v1/internal/users/batch`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ ids: unique }),
                signal: controller.signal
            })

            if (!res.ok) {
                throw new Error(`user service returned ${res.status}`)
            }

            const body = (await res.json()) as BatchResponse
            return new Map((body.data ?? []).map((identity) => [identity.id, identity]))
        } catch (err) {
            logger.error({ err }, 'failed to resolve user identities')
            throw err
        } finally {
            clearTimeout(timeout)
        }
    }
}
