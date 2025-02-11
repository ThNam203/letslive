import WebSocket, { WebSocketServer } from 'ws'
import { Message } from './models/Message'
import { RedisService } from './services/redis'
import { ChatEventType } from './types/chat_event'
import { ChatMessage, ChatMessageType } from './types/chat_message'

type UserInfo = { currentRoom: string | null; id: string | null; name: string | null }

export class ChatServer {
    private connections: Map<string, WebSocket> = new Map()

    constructor(
        private redisService: RedisService,
        private messageModel: typeof Message,
        private wss: WebSocketServer
    ) {
        this.initialize()
    }

    public getConnections() {
        return this.connections
    }

    private initialize() {
        this.wss.on('connection', (ws) => this.handleConnection(ws))
        this.redisService.subscribeToMessages((pattern, channel, message) => this.handleRedisMessage(channel, message))
    }

    private async handleConnection(ws: WebSocket) {
        let userInfo: UserInfo = {
            currentRoom: null,
            id: null,
            name: null
        }

        ws.on('message', async (rawMessage) => {
            let data: ChatMessage
            try {
                data = JSON.parse(rawMessage.toString())
                userInfo = {
                    currentRoom: data.room,
                    id: data.senderId,
                    name: data.senderName
                }
                this.connections.set(userInfo.id!, ws)
            } catch (err) {
                console.error(err)
                return
            }

            await this.handleWebSocketMessage(data, userInfo)
        })

        ws.on('close', () => {
            if (userInfo.currentRoom) {
                this.redisService.removeUserFromRoom(userInfo.id!, userInfo.currentRoom)
                this.redisService.publishEvent(userInfo.currentRoom, ChatEventType.LEAVE, userInfo.id!, userInfo.name!)
            }

            this.connections.delete(userInfo.id!)
            userInfo = {
                currentRoom: null,
                id: null,
                name: null
            }
        })
    }

    private async handleWebSocketMessage(data: ChatMessage, userInfo: UserInfo) {
        switch (data.type) {
            case ChatMessageType.JOIN:
                await this.handleJoin(data, userInfo)
                break
            case ChatMessageType.LEAVE:
                await this.handleLeave(data, userInfo)
                break
            case ChatMessageType.MESSAGE:
                await this.handleMessage(data, userInfo)
                break
        }
    }

    private async handleJoin(data: ChatMessage, userInfo: UserInfo) {
        if (userInfo.currentRoom === data.room && (await this.redisService.checkIfUserInRoom(data.senderId, data.room)))
            return

        if (userInfo.currentRoom && (await this.redisService.checkIfUserInRoom(data.senderId, userInfo.currentRoom))) {
            await this.redisService.removeUserFromRoom(data.senderId, userInfo.currentRoom)
        }

        userInfo.currentRoom = data.room
        await this.redisService.addUserToRoom(data.senderId, data.room)
        await this.redisService.publishEvent(data.room, ChatEventType.JOIN, data.senderId, data.senderName)
    }

    private async handleLeave(data: ChatMessage, userInfo: UserInfo) {
        if (
            !userInfo.currentRoom ||
            userInfo.currentRoom !== data.room ||
            !this.redisService.checkIfUserInRoom(data.senderId, userInfo.currentRoom)
        ) {
            console.error('Something went wrong')
            return
        }

        await this.redisService.removeUserFromRoom(data.senderId, userInfo.currentRoom)
        await this.redisService.publishEvent(data.room, ChatEventType.LEAVE, data.senderId, data.senderName)
        userInfo.currentRoom = null
        this.connections.delete(data.senderId)
    }

    private async handleMessage(data: ChatMessage, userInfo: UserInfo) {
        if (
            !userInfo.currentRoom ||
            userInfo.currentRoom !== data.room ||
            !this.redisService.checkIfUserInRoom(data.senderId, userInfo.currentRoom)
        ) {
            console.error('Something went wrong')
            return
        }

        await this.redisService.publishMessage(data.room, data)
        await new this.messageModel(data).save()
    }

    private async handleRedisMessage(channel: string, message: string) {
        const roomId = channel.split(':')[1]
        const roomMembers = await this.redisService.getUsersInRoom(roomId)

        roomMembers.forEach((memberId) => {
            const connection = this.connections.get(memberId)
            if (connection) {
                connection.send(message)
            }
        })
    }
}
