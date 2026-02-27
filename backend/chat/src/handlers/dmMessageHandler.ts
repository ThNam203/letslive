import { Request, Response } from 'express'
import { DmMessageService } from '../services/dmMessageService'
import { RESPONSE_TEMPLATES, newResponseFromTemplate, Response as ServiceResponse } from '../types/api-response'
import { SendDmMessageRequest, EditDmMessageRequest } from '../types/conversation'

function writeResponse(req: Request, res: Response, resData: ServiceResponse<any>) {
    resData.requestId = req.requestId ?? ''
    res.status(resData.statusCode).json(resData)
}

export class DmMessageHandler {
    constructor(private dmMessageService: DmMessageService) {}

    getMessages = async (req: Request, res: Response) => {
        const userId = req.userId!
        const conversationId = req.params.id
        const before = req.query.before as string | undefined
        const limit = Math.min(100, Math.max(1, parseInt(req.query.limit as string) || 50))

        const result = await this.dmMessageService.getMessages(conversationId, userId, before, limit)
        writeResponse(req, res, result)
    }

    sendMessage = async (req: Request, res: Response) => {
        const userId = req.userId!
        const conversationId = req.params.id
        const body = req.body as SendDmMessageRequest

        if (!body.text || typeof body.text !== 'string') {
            writeResponse(req, res, newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT))
            return
        }

        if (body.type && !['text', 'image'].includes(body.type)) {
            writeResponse(req, res, newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT))
            return
        }

        const senderUsername = req.body.senderUsername || userId

        const result = await this.dmMessageService.sendMessage(
            conversationId,
            userId,
            senderUsername,
            body.text,
            body.type || 'text',
            body.imageUrls,
            body.replyTo
        )

        writeResponse(req, res, result)
    }

    editMessage = async (req: Request, res: Response) => {
        const userId = req.userId!
        const conversationId = req.params.id
        const messageId = req.params.msgId
        const body = req.body as EditDmMessageRequest

        if (!body.text || typeof body.text !== 'string') {
            writeResponse(req, res, newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT))
            return
        }

        const result = await this.dmMessageService.editMessage(conversationId, messageId, userId, body.text)
        writeResponse(req, res, result)
    }

    deleteMessage = async (req: Request, res: Response) => {
        const userId = req.userId!
        const conversationId = req.params.id
        const messageId = req.params.msgId

        const result = await this.dmMessageService.deleteMessage(conversationId, messageId, userId)
        writeResponse(req, res, result)
    }

    markAsRead = async (req: Request, res: Response) => {
        const userId = req.userId!
        const conversationId = req.params.id
        const messageId = req.body.messageId as string | undefined

        const result = await this.dmMessageService.markAsRead(conversationId, userId, messageId)
        writeResponse(req, res, result)
    }
}
