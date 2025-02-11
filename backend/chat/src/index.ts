import Redis from 'ioredis'
import mongoose from 'mongoose'
import { WebSocketServer } from 'ws'
import { ChatServer } from './chatServer'
import { Message } from './models/Message'
import { RedisService } from './services/redis'
import esMain from 'es-main'

async function createServer() {
    const pub = new Redis(6379, 'localhost')
    const sub = new Redis(6379, 'localhost')
    const roomManager = new Redis(6379, 'localhost')

    await mongoose.connect('mongodb://localhost:27017/chat')

    const redisService = new RedisService(pub, sub, roomManager)
    const wss = new WebSocketServer({ port: 8080 })

    return new ChatServer(redisService, Message, wss)
}

// no need actually
if (esMain(import.meta)) {
    createServer()
        .then(() => console.log('Server started'))
        .catch(console.error)
}
