import Consul from 'consul'
import { RegisterOptions } from 'consul/lib/agent/service'
import crypto from 'crypto'

export type ConsulConfig = {
    serviceName: string
    hostname: string
    port: number
    healthCheckURL: string
}

class ConsulRegistry {
    private client: Consul
    private config: ConsulConfig
    private instanceId: string
    private isRegistered: boolean

    constructor(consulHost: string, consulPort: number, config: ConsulConfig) {
        this.client = new Consul({ host: consulHost, port: consulPort })
        this.config = config
        this.instanceId = ''
        this.isRegistered = false
    }

    async register() {
        this.instanceId = this.generateInstanceId()

        const registerOptions = {
            id: this.instanceId,
            name: this.config.serviceName,
            address: this.config.hostname,
            port: this.config.port,
            check: {
                http: this.config.healthCheckURL,
                interval: '15s',
                timeout: '2s'
            }
        } as RegisterOptions

        const maxRetries = 5

        for (let i = 0; i < maxRetries; i++) {
            try {
                await this.client.agent.service.register(registerOptions)
                this.isRegistered = true
                return
            } catch (err) {
                if (i === maxRetries - 1) {
                    throw err
                } else {
                    await new Promise((resolve) => setTimeout(resolve, 5000))
                }
            }
        }
    }

    async deregister() {
        if (this.isRegistered) await this.client.agent.service.deregister(this.instanceId)
    }

    async serviceAddresses(serviceName: string) {
        const services = await this.client.health.service({ service: serviceName, passing: true })

        if (!services.length) {
            throw new Error('Service not found')
        }

        return services.map((entry) => `${entry.Service.Address}:${entry.Service.Port}`)
    }

    async serviceAddress(serviceName: string) {
        const addrs = await this.serviceAddresses(serviceName)
        return addrs[Math.floor(Math.random() * addrs.length)]
    }

    private generateInstanceId() {
        const maxSafeLimit = 281474976710655;
        return this.config.serviceName + '-' + crypto.randomInt(10000000, maxSafeLimit)
    }
}

export default ConsulRegistry
