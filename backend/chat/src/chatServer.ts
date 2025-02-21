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
                    currentRoom: data.roomId,
                    id: data.userId,
                    name: data.username
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

        // if the proxy doesn't get an upstream response in 60s (kong default), it will close the connection
        // the ws.send() does not count as an upstream response
        setInterval(() => {
            ws.ping()
        }, 30000)
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
        if (
            userInfo.currentRoom === data.roomId &&
            (await this.redisService.checkIfUserInRoom(data.userId, data.roomId))
        ) {
            return
        }

        if (userInfo.currentRoom && (await this.redisService.checkIfUserInRoom(data.userId, userInfo.currentRoom))) {
            await this.redisService.removeUserFromRoom(data.userId, userInfo.currentRoom)
        }

        userInfo.currentRoom = data.roomId
        await this.redisService.addUserToRoom(data.userId, data.roomId)
        await this.redisService.publishEvent(data.roomId, ChatEventType.JOIN, data.userId, data.username)
    }

    private async handleLeave(data: ChatMessage, userInfo: UserInfo) {
        if (
            !userInfo.currentRoom ||
            userInfo.currentRoom !== data.roomId ||
            !this.redisService.checkIfUserInRoom(data.userId, userInfo.currentRoom)
        ) {
            console.error('Something went wrong')
            return
        }

        await this.redisService.removeUserFromRoom(data.userId, userInfo.currentRoom)
        await this.redisService.publishEvent(data.roomId, ChatEventType.LEAVE, data.userId, data.username)
        userInfo.currentRoom = null
        this.connections.delete(data.userId)
    }

    private async handleMessage(data: ChatMessage, userInfo: UserInfo) {
        if (
            !userInfo.currentRoom ||
            userInfo.currentRoom !== data.roomId ||
            !this.redisService.checkIfUserInRoom(data.userId, userInfo.currentRoom)
        ) {
            console.error('Something went wrong')
            return
        }

        await this.redisService.publishMessage(data.roomId, data)
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
