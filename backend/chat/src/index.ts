import Redis from 'ioredis'
import mongoose from 'mongoose'
import { WebSocketServer } from 'ws'
import { ChatServer } from './chatServer'
import { Message } from './models/Message'
import { RedisService } from './services/redis'
import esMain from 'es-main'
import { createServer, Server } from 'http'
import ConsulRegistry from 'services/discovery'
import { RESPONSE_TEMPLATES, Response as ServiceResponse, newResponseFromTemplate } from './types/api-response'
import express, { NextFunction, Request, Response } from 'express'
import requestIdMiddleware from 'middlewares/requestId'
import loggingMiddleware from 'middlewares/logging'
import pinohttp from 'pino-http'
import logger from 'lib/logger'

function asyncHandler(fn: (req: Request, res: Response, next: NextFunction) => Promise<any>) {
    return (req: Request, res: Response, next: NextFunction) => {
        fn(req, res, next).catch(next) // Pass errors to the error-handling middleware
    }
}

function CreateExpressServer() {
    const app = express()
    const pinoM = pinohttp({
        logger: logger,
        autoLogging: {
            ignore: (req: Request) => {
                return req.url === '/v1/health' || req.method === 'OPTIONS'
            }
        }
    })

    app.use(pinoM)
    app.use(requestIdMiddleware)
    // app.use(loggingMiddleware) replaced by pinohttp

    app.get('/v1/health', (req, res) => {
        res.json({ status: 'ok' })
    })

    // TODO: update response
    app.get(
        '/v1/messages',
        asyncHandler(async (req, res) => {
            const roomId = req.query.roomId as string
            if (!roomId || roomId.length > 36) {
                writeResponse(req, res, newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_ROOM_NOT_FOUND))
                return
            }

            const messages = await Message.find({ roomId }).sort({ timestamp: 1 }).limit(50)
            // TODO: shouldn't be any
            writeResponse(req, res, newResponseFromTemplate<any>(RESPONSE_TEMPLATES.RES_SUCC_OK, messages))
        })
    )

    app.get('*', (req: Request, res: Response) => {
        writeResponse(req, res, newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_ROUTE_NOT_FOUND))
    })

    app.use((err: any, req: Request, res: Response, next: NextFunction): void => {
        logger.error(err) // Log the error for debugging
        writeResponse(req, res, newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_ROUTE_NOT_FOUND))
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
        .then(() => logger.info('Server started'))
        .catch(logger.error)

    const consul = CreateConsulRegistry()
    consul.register()

    server.listen('7780', () => {
        logger.info(`Server started on port ${'7780'}`)
    })

    process.on('SIGINT', () => shutdown('SIGINT', consul))
    process.on('SIGTERM', () => shutdown('SIGTERM', consul))

    process.on('uncaughtException', (error) => {
        logger.error(`Uncaught Exception: ${error.message}`)
        process.exit(1)
    })

    process.on('unhandledRejection', (reason, promise) => {
        logger.error({
            message: 'Unhandled Rejection',
            reason: reason instanceof Error ? reason.stack : reason
        })
        process.exit(1)
    })
}

const shutdown = async (signal: string, consul: ConsulRegistry) => {
    logger.info(`\nReceived ${signal}. Starting graceful shutdown...`)
    try {
        logger.info('Deregistering from Consul...')
        if (consul && typeof consul.deregister === 'function') {
            await consul.deregister()
            logger.info('successfully deregistered from Consul.')
        } else {
            logger.warn('consul client or deregister function not available.')
        }
        process.exit(0)
    } catch (err) {
        logger.error(`error during ${signal} cleanup: ${err}`)
        process.exit(1)
    }
}

const writeResponse = function (req: Request, res: Response, resData: ServiceResponse<any>) {
    resData.requestId = req.requestId ?? ''
    res.status(resData.statusCode).json(resData)
}
