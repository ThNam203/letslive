import Redis from 'ioredis'
import mongoose from 'mongoose'
import { WebSocketServer } from 'ws'
import { ChatServer } from './chatServer'
import { Message } from './models/Message'
import { RedisService } from './services/redis'
import esMain from 'es-main'
import { createServer, Server } from 'http'
import ConsulRegistry from 'services/discovery'
import { ApiErrors } from 'types/api_error'
import express, { NextFunction, Request, Response } from 'express'
import requestIdMiddleware from 'middlewares/requestId'
import loggingMiddleware from 'middlewares/logging'

function asyncHandler(fn: (req: Request, res: Response, next: NextFunction) => Promise<any>) {
    return (req: Request, res: Response, next: NextFunction) => {
        fn(req, res, next).catch(next) // Pass errors to the error-handling middleware
    }
}

function CreateExpressServer() {
    const app = express()

    app.use(requestIdMiddleware)
    app.use(loggingMiddleware)

    app.get('/v1/health', (req, res) => {
        res.json({ status: 'ok' })
    })

    app.get(
        '/v1/messages',
        asyncHandler(async (req, res) => {
            const roomId = req.query.roomId as string
            if (!roomId) {
                res.status(400).json(ApiErrors.INVALID_PATH)
                return
            }

            const messages = await Message.find({ roomId }).sort({ timestamp: 1 }).limit(50)
            res.json(messages)
        })
    )

    app.get('*', (_, res: Response) => {
        res.status(404).json(ApiErrors.ROUTE_NOT_FOUND)
    })

    app.use((err: any, req: Request, res: Response, next: NextFunction): void => {
        console.error(err) // Log the error for debugging
        res.status(500).json(ApiErrors.INTERNAL_SERVER_ERROR)
    })

    return createServer(app)
}

async function SetupWebSocketServer(server: Server) {
    const pub = new Redis(6379, 'chat_pubsub')
    const sub = new Redis(6379, 'chat_pubsub')
    const roomManager = new Redis(6379, 'chat_pubsub')

    await mongoose.connect(
        `mongodb://${process.env.CHAT_DB_USER}:${process.env.CHAT_DB_PASSWORD}@chat_db:27017/chat?authSource=admin`
    )

    const redisService = new RedisService(pub, sub, roomManager)
    const wss = new WebSocketServer({ server, path: '/v1/ws' })

    return new ChatServer(redisService, Message, wss)
}

// TODO: add config instead of hard-coded
function CreateConsulRegistry() {
    return new ConsulRegistry('consul', 8500, {
        serviceName: 'chat',
        hostname: 'chat',
        port: 7780,
        healthCheckURL: 'http://chat:7780/v1/health'
    })
}

// no need actually
if (esMain(import.meta)) {
    const server = CreateExpressServer()

    SetupWebSocketServer(server)
        .then(() => console.log('Server started'))
        .catch(console.error)

    const consul = CreateConsulRegistry()
    consul.register()

    server.listen('7780', () => {
        console.log(`Server started on port ${'7780'}`)
    })

    process.on('SIGINT', () => shutdown('SIGINT', consul))
    process.on('SIGTERM', () => shutdown('SIGTERM', consul))

    process.on('uncaughtException', (error) => {
        console.error('Uncaught Exception:', error)
        process.exit(1)
    })

    process.on('unhandledRejection', (reason, promise) => {
        console.error('Unhandled Rejection at:', promise, 'reason:', reason)
        process.exit(1)
    })
}

const shutdown = async (signal: string, consul: ConsulRegistry) => {
    console.log(`\nReceived ${signal}. Starting graceful shutdown...`)
    try {
        console.log('Deregistering from Consul...')
        if (consul && typeof consul.deregister === 'function') {
            await consul.deregister()
            console.log('successfully deregistered from Consul.')
        } else {
            console.warn('consul client or deregister function not available.')
        }
        process.exit(0)
    } catch (err) {
        console.error(`error during ${signal} cleanup:`, err)
        process.exit(1)
    }
}