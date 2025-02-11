import WebSocket, { WebSocketServer } from 'ws'
import { MongoMemoryServer } from 'mongodb-memory-server'
import mongoose from 'mongoose'
import Redis from 'ioredis-mock'
import { Message } from '../models/Message'
import { ChatMessage, ChatMessageType } from '../types/chat_message'
import { ChatServer } from '../chatServer'
import { RedisService } from '../services/redis'
import { log } from 'console'
import { ChatEvent } from '../types/chat_event'

describe('test join and leave events', () => {
    const serverPort = 8080
    let mongoServer: MongoMemoryServer
    let chatServer: ChatServer
    let redisService: RedisService
    let wss: WebSocketServer

    beforeAll(async () => {
        // Setup MongoDB Memory Server
        mongoServer = await MongoMemoryServer.create()
        const mongoUri = mongoServer.getUri()
        await mongoose.connect(mongoUri)

        // Setup Redis Mock
        const pub = new Redis()
        const sub = new Redis()
        const roomManager = new Redis()
        redisService = new RedisService(pub, sub, roomManager)

        // Setup WebSocket Server
        wss = new WebSocketServer({ port: serverPort })

        chatServer = new ChatServer(redisService, Message, wss)
    })

    afterAll(async () => {
        await mongoose.disconnect()
        await mongoServer.stop()
        wss.close()
    })

    afterEach(async () => {
        await Message.deleteMany({})
        await new Redis().flushall()
    })

    it('should handle user joining a room', async () => {
        const ws = new WebSocket(`ws://localhost:${serverPort}`)
        await new Promise((resolve) => ws.on('open', resolve))

        const message = {
            type: ChatMessageType.JOIN,
            room: 'test-room',
            senderId: 'user1',
            senderName: 'Test User'
        }

        ws.send(JSON.stringify(message))

        // Wait for the join event to be processed
        await new Promise((resolve) => setTimeout(resolve, 50))

        const usersInRoom = await redisService.getUsersInRoom('test-room')
        expect(usersInRoom).toContain('user1')

        expect(chatServer.getConnections().get('user1')).toBeDefined()

        ws.close()
    })

    it('should handle users join a room and leave (emitting events)', async () => {
        const ws = new WebSocket(`ws://localhost:${serverPort}`)
        const ws2 = new WebSocket(`ws://localhost:${serverPort}`)

        await new Promise((resolve) => ws.on('open', resolve))
        await new Promise((resolve) => ws2.on('open', resolve))

        const joinMessage1 = {
            type: ChatMessageType.JOIN,
            room: 'test-room',
            senderId: 'user1',
            senderName: 'Test User'
        }

        const joinMessage2 = {
            type: ChatMessageType.JOIN,
            room: 'test-room',
            senderId: 'user2',
            senderName: 'Test User 2'
        }

        ws.send(JSON.stringify(joinMessage1))
        ws2.send(JSON.stringify(joinMessage2))

        // Wait for the join event to be processed
        await new Promise((resolve) => setTimeout(resolve, 50))

        const usersInRoom = await redisService.getUsersInRoom('test-room')
        expect(usersInRoom).toHaveLength(2)
        expect(usersInRoom).toEqual(expect.arrayContaining(['user1', 'user2']))

        expect(chatServer.getConnections().get('user1')).toBeDefined()
        expect(chatServer.getConnections().get('user2')).toBeDefined()

        const leaveMessage1 = {
            type: ChatMessageType.LEAVE,
            room: 'test-room',
            senderId: 'user1',
            senderName: 'Test User'
        }

        const leaveMessage2 = {
            type: ChatMessageType.LEAVE,
            room: 'test-room',
            senderId: 'user2',
            senderName: 'Test User 2'
        }

        ws.send(JSON.stringify(leaveMessage1))
        ws2.send(JSON.stringify(leaveMessage2))

        // Wait for the leave event to be processed
        await new Promise((resolve) => setTimeout(resolve, 50))

        const usersLeftInRoom = await redisService.getUsersInRoom('test-room')
        expect(usersLeftInRoom).toHaveLength(0)

        expect(chatServer.getConnections().get('user1')).not.toBeDefined()
        expect(chatServer.getConnections().get('user2')).not.toBeDefined()

        ws.close()
        ws2.close()
    })

    it('should handle join and leave unexpectedly without emitting leave event', async () => {
        const ws = new WebSocket(`ws://localhost:${serverPort}`)

        await new Promise((resolve) => ws.on('open', resolve))
        let times = 0

        ws.on('message', async (rawData) => {
            if (times === 0) {
                const data: ChatMessage = JSON.parse(rawData.toString())
                expect(data.type).toBe(ChatMessageType.JOIN)

                const usersInRoom = await redisService.getUsersInRoom('test-room')
                expect(usersInRoom).toContain('user1')
                expect(chatServer.getConnections().get('user1')).toBeDefined()
                times++
            } else if (times === 1) {
                const data: ChatMessage = JSON.parse(rawData.toString())
                expect(data.type).toBe(ChatMessageType.LEAVE)
                const usersInRoom = await redisService.getUsersInRoom('test-room')
                expect(usersInRoom).not.toContain('user1')
                expect(chatServer.getConnections().get('user1')).not.toBeDefined()
                times++
            } else {
                throw new Error('Unexpected message')
            }
        })

        // First join the room
        const joinMessage = {
            type: ChatMessageType.JOIN,
            room: 'test-room',
            senderId: 'user1',
            senderName: 'Test User'
        }

        ws.send(JSON.stringify(joinMessage))

        await new Promise((resolve) => setTimeout(resolve, 50))

        const leaveMessage = {
            type: ChatMessageType.LEAVE,
            room: 'test-room',
            senderId: 'user1',
            senderName: 'Test User'
        }

        ws.send(JSON.stringify(leaveMessage))

        // Wait for the join and leave events to be processed
        await new Promise((resolve) => setTimeout(resolve, 50))

        ws.close()
    })

    it('should handle users join a room and leave', async () => {
        const ws = new WebSocket(`ws://localhost:${serverPort}`)
        const ws2 = new WebSocket(`ws://localhost:${serverPort}`)

        await new Promise((resolve) => ws.on('open', resolve))
        await new Promise((resolve) => ws2.on('open', resolve))

        const joinMessage1 = {
            type: ChatMessageType.JOIN,
            room: 'test-room',
            senderId: 'user1',
            senderName: 'Test User'
        }

        const joinMessage2 = {
            type: ChatMessageType.JOIN,
            room: 'test-room',
            senderId: 'user2',
            senderName: 'Test User 2'
        }

        ws.send(JSON.stringify(joinMessage1))
        ws2.send(JSON.stringify(joinMessage2))

        // Wait for the join event to be processed
        await new Promise((resolve) => setTimeout(resolve, 50))

        const usersInRoom = await redisService.getUsersInRoom('test-room')
        expect(usersInRoom).toHaveLength(2)
        expect(usersInRoom).toEqual(expect.arrayContaining(['user1', 'user2']))

        expect(chatServer.getConnections().get('user1')).toBeDefined()
        expect(chatServer.getConnections().get('user2')).toBeDefined()

        ws.close()
        ws2.close()

        // Wait for the leave event to be processed
        await new Promise((resolve) => setTimeout(resolve, 50))

        const usersLeftInRoom = await redisService.getUsersInRoom('test-room')
        expect(usersLeftInRoom).toHaveLength(0)

        expect(chatServer.getConnections().get('user1')).not.toBeDefined()
        expect(chatServer.getConnections().get('user2')).not.toBeDefined()
    })

    it('should handle users join multiple times (without leaving)', async () => {
        const ws = new WebSocket(`ws://localhost:${serverPort}`)

        await new Promise((resolve) => ws.on('open', resolve))

        const joinMessage = {
            type: ChatMessageType.JOIN,
            room: 'test-room',
            senderId: 'user1',
            senderName: 'Test User'
        }

        ws.send(JSON.stringify(joinMessage))
        ws.send(JSON.stringify(joinMessage))
        ws.send(JSON.stringify(joinMessage))
        ws.send(JSON.stringify(joinMessage))

        // Wait for the join event to be processed
        await new Promise((resolve) => setTimeout(resolve, 50))

        const usersInRoom = await redisService.getUsersInRoom('test-room')
        expect(usersInRoom).toHaveLength(1)
        expect(usersInRoom).toContain('user1')

        expect(chatServer.getConnections().get('user1')).toBeDefined()
        ws.close()
    })

    it('should handle users join a room and leave multiple times (without joining again)', async () => {
        const ws = new WebSocket(`ws://localhost:${serverPort}`)

        await new Promise((resolve) => ws.on('open', resolve))

        const joinMessage = {
            type: ChatMessageType.JOIN,
            room: 'test-room',
            senderId: 'user1',
            senderName: 'Test User'
        }

        ws.send(JSON.stringify(joinMessage))

        await new Promise((resolve) => setTimeout(resolve, 50))

        const leaveMessage = {
            type: ChatMessageType.LEAVE,
            room: 'test-room',
            senderId: 'user1',
            senderName: 'Test User'
        }

        ws.send(JSON.stringify(leaveMessage))
        ws.send(JSON.stringify(leaveMessage))
        ws.send(JSON.stringify(leaveMessage))
        ws.send(JSON.stringify(leaveMessage))
        ws.send(JSON.stringify(leaveMessage))

        await new Promise((resolve) => setTimeout(resolve, 50))

        const usersLeftInRoom = await redisService.getUsersInRoom('test-room')
        expect(usersLeftInRoom).toHaveLength(0)

        expect(chatServer.getConnections().get('user1')).not.toBeDefined()

        ws.close()
    })
})

describe('test messaging', () => {
    const serverPort = 8080
    let mongoServer: MongoMemoryServer
    let chatServer: ChatServer
    let redisService: RedisService
    let wss: WebSocketServer

    beforeAll(async () => {
        // Setup MongoDB Memory Server
        mongoServer = await MongoMemoryServer.create()
        const mongoUri = mongoServer.getUri()
        await mongoose.connect(mongoUri)

        // Setup Redis Mock
        const pub = new Redis()
        const sub = new Redis()
        const roomManager = new Redis()
        redisService = new RedisService(pub, sub, roomManager)

        // Setup WebSocket Server
        wss = new WebSocketServer({ port: serverPort })

        chatServer = new ChatServer(redisService, Message, wss)
    })

    afterAll(async () => {
        await mongoose.disconnect()
        await mongoServer.stop()
        wss.close()
    })

    afterEach(async () => {
        await Message.deleteMany({})
        await new Redis().flushall()
    })

    it('server must receive and client gets its own message', async () => {
        const ws = new WebSocket(`ws://localhost:${serverPort}`)

        await new Promise((resolve) => ws.on('open', resolve))

        ws.on('message', (rawData) => {
            const data: ChatMessage = JSON.parse(rawData.toString())

            if (data.type === ChatMessageType.JOIN) return
            if (data.type === ChatMessageType.LEAVE)
                throw new Error('user cant receive its own leave message cause it left lol')
            if (data.type === ChatMessageType.MESSAGE) {
                expect(data.text).toBe('Hello World')
                expect(data.senderId).toBe('user1')
                expect(data.senderName).toBe('Test User')
                expect(data.room).toBeUndefined()
            }
        })

        const joinMessage = {
            type: ChatMessageType.JOIN,
            room: 'test-room',
            senderId: 'user1',
            senderName: 'Test User'
        }

        ws.send(JSON.stringify(joinMessage))

        await new Promise((resolve) => setTimeout(resolve, 50))

        const message = {
            type: ChatMessageType.MESSAGE,
            room: 'test-room',
            senderId: 'user1',
            senderName: 'Test User',
            text: 'Hello World'
        }

        ws.send(JSON.stringify(message))
        await new Promise((resolve) => setTimeout(resolve, 50))

        const leaveMessage = {
            type: ChatMessageType.LEAVE,
            room: 'test-room',
            senderId: 'user1',
            senderName: 'Test User'
        }

        ws.send(JSON.stringify(leaveMessage))

        await new Promise((resolve) => setTimeout(resolve, 50))

        ws.close()
    })

    it(`other client receives others's messages`, async () => {
        const ws = new WebSocket(`ws://localhost:${serverPort}`)
        const ws2 = new WebSocket(`ws://localhost:${serverPort}`)

        await new Promise((resolve) => ws.on('open', resolve))
        await new Promise((resolve) => ws2.on('open', resolve))

        ws.on('message', (rawData) => {
            const dataType: { type: string } = JSON.parse(rawData.toString())

            if (dataType.type === ChatMessageType.JOIN) {
                const event: ChatEvent = JSON.parse(rawData.toString())

                for (const [idx, user] of ['user1', 'user2'].entries()) {
                    if (event.userId === user) {
                        expect(event.userId).toBe(user)
                        expect(event.username).toBe(`Test User ${idx + 1}`)
                    }
                }
            }
            if (dataType.type === ChatMessageType.LEAVE) {
                const event: ChatEvent = JSON.parse(rawData.toString())
                if (event.userId === 'user1') throw new Error('user should not receive its own leave message')

                expect(event.userId).toBe('user2')
                expect(event.username).toBe(`Test User 2`)
            }
            if (dataType.type === ChatMessageType.MESSAGE) {
                const data: ChatMessage = JSON.parse(rawData.toString())

                expect(data.text).toBe('Hello World')
                expect(data.senderId).toBe('user2')
                expect(data.senderName).toBe('Test User 2')
                expect(data.room).toBeUndefined()
            }
        })

        const joinMessage = {
            type: ChatMessageType.JOIN,
            room: 'test-room',
            senderId: 'user1',
            senderName: 'Test User 1'
        }

        const joinMessage2 = {
            type: ChatMessageType.JOIN,
            room: 'test-room',
            senderId: 'user2',
            senderName: 'Test User 2'
        }

        ws.send(JSON.stringify(joinMessage))
        ws2.send(JSON.stringify(joinMessage2))

        await new Promise((resolve) => setTimeout(resolve, 50))

        const message = {
            type: ChatMessageType.MESSAGE,
            room: 'test-room',
            senderId: 'user2',
            senderName: 'Test User 2',
            text: 'Hello World'
        }

        ws2.send(JSON.stringify(message))
        await new Promise((resolve) => setTimeout(resolve, 50))

        const leaveMessage = {
            type: ChatMessageType.LEAVE,
            room: 'test-room',
            senderId: 'user2',
            senderName: 'Test User 2'
        }

        ws2.send(JSON.stringify(leaveMessage))

        await new Promise((resolve) => setTimeout(resolve, 50))

        ws.close()
        ws2.close()
    })
})
