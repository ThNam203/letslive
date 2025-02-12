import Redis from 'ioredis'
import mongoose from 'mongoose'
import { WebSocketServer } from 'ws'
import { ChatServer } from './chatServer'
import { Message } from './models/Message'
import { RedisService } from './services/redis'
import esMain from 'es-main'

async function createServer() {
    const pub = new Redis(6379, 'chat_pubsub')
    const sub = new Redis(6379, 'chat_pubsub')
    const roomManager = new Redis(6379, 'chat_pubsub')

    await mongoose.connect('mongodb://chat_db:27017/chat')

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
