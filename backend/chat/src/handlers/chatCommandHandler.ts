import { Request, Response } from 'express'
import { ChatCommandService, ChatCommandScope } from '../services/chatCommandService'
import { extractUserIdFromCookie } from '../middlewares/auth'
import { RESPONSE_TEMPLATES, newResponseFromTemplate, Response as ServiceResponse } from '../types/api-response'

function writeResponse(req: Request, res: Response, resData: ServiceResponse<any>) {
    resData.requestId = req.requestId ?? ''
    res.status(resData.statusCode).json(resData)
}

export class ChatCommandHandler {
    constructor(private chatCommandService: ChatCommandService) {}

    listForRoom = async (req: Request, res: Response) => {
        const roomId = (req.query.roomId as string) || ''
        if (!roomId || roomId.length > 36) {
            writeResponse(req, res, newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_ROOM_NOT_FOUND))
            return
        }
        const userId = extractUserIdFromCookie(req.headers.cookie)
        const result = await this.chatCommandService.listForRoom(roomId, userId)
        writeResponse(req, res, result)
    }

    listMine = async (req: Request, res: Response) => {
        const result = await this.chatCommandService.listMine(req.userId!)
        writeResponse(req, res, result)
    }

    create = async (req: Request, res: Response) => {
        const { scope, name, response, description } = req.body ?? {}
        const result = await this.chatCommandService.create(
            req.userId!,
            scope as ChatCommandScope,
            name,
            response,
            description ?? ''
        )
        writeResponse(req, res, result)
    }

    update = async (req: Request, res: Response) => {
        const id = req.params.id
        if (!id || id.length > 36) {
            writeResponse(req, res, newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT))
            return
        }
        const { name, response, description } = req.body ?? {}
        const result = await this.chatCommandService.update(
            req.userId!,
            id,
            name,
            response,
            description
        )
        writeResponse(req, res, result)
    }

    delete = async (req: Request, res: Response) => {
        const id = req.params.id
        if (!id || id.length > 36) {
            writeResponse(req, res, newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT))
            return
        }
        const result = await this.chatCommandService.delete(req.userId!, id)
        writeResponse(req, res, result)
    }
}
