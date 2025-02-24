import Redis from 'ioredis'
import mongoose from 'mongoose'
import { WebSocketServer } from 'ws'
import { ChatServer } from './chatServer'
import { Message } from './models/Message'
import { RedisService } from './services/redis'
import esMain from 'es-main'
import express from 'express'
import { createServer, Server } from 'http'
import ConsulRegistry from 'services/discovery'

function CreateExpressServer() {
    const app = express()

    app.get('/v1/health', (req, res) => {
        res.json({ status: 'ok' })
    })

    app.get('/v1/messages', async (req, res) => {
        const roomId = req.query.roomId as string
        if (!roomId) {
            res.status(400).json({ error: 'roomId is required' })
            return
        }

        const messages = await Message.find({ roomId }).sort({ timestamp: -1 }).limit(50)
        res.json(messages)
    })

    return createServer(app)
}

async function SetupWebSocketServer(server: Server) {
    const pub = new Redis(6379, 'chat_pubsub')
    const sub = new Redis(6379, 'chat_pubsub')
    const roomManager = new Redis(6379, 'chat_pubsub')

    await mongoose.connect('mongodb://chat_db:27017/chat')

    const redisService = new RedisService(pub, sub, roomManager)
    const wss = new WebSocketServer({ server })

    return new ChatServer(redisService, Message, wss)
}

// TODO: add config instead of hard-coded
function CreateConsulRegistry() {
    return new ConsulRegistry('consul', 8500, {
        serviceName: 'chat',
        hostname: 'chat',
        port: 8080,
        healthCheckURL: 'http://chat:8080/v1/health'
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

    server.listen('8080', () => {
        console.log(`Server started on port ${'8080'}`)
    })

    process.on('SIGINT', async () => {
        await consul.deregister()
        process.exit(0)
    })
}
