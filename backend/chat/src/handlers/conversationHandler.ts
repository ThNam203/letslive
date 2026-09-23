import { Request, Response } from 'express'
import { ConversationService } from '../services/conversationService'
import { UserServiceGateway } from '../gateway/userService'
import { RESPONSE_TEMPLATES, newResponseFromTemplate, Response as ServiceResponse } from '../types/api-response'
import {
    CreateConversationRequest,
    UpdateConversationRequest,
    AddParticipantRequest,
    ConversationType
} from '../types/conversation'

function writeResponse(req: Request, res: Response, resData: ServiceResponse<any>) {
    resData.requestId = req.requestId ?? ''
    res.status(resData.statusCode).json(resData)
}

export class ConversationHandler {
    constructor(
        private conversationService: ConversationService,
        private userServiceGateway: UserServiceGateway
    ) {}

    createConversation = async (req: Request, res: Response) => {
        const userId = req.userId!
        const body = req.body as CreateConversationRequest

        if (!body.type || !Object.values(ConversationType).includes(body.type)) {
            writeResponse(req, res, newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT))
            return
        }

        if (!body.participantIds || !Array.isArray(body.participantIds) || body.participantIds.length === 0) {
            writeResponse(req, res, newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT))
            return
        }

        // Validate all participant IDs are strings and have valid length
        for (const id of body.participantIds) {
            if (typeof id !== 'string' || id.length > 36) {
                writeResponse(req, res, newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT))
                return
            }
        }

        // Identities come from the user service, never from the request body:
        // the creator would otherwise get to choose what everyone is called.
        let identities
        try {
            identities = await this.userServiceGateway.getIdentities([userId, ...body.participantIds])
        } catch {
            writeResponse(req, res, newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_INTERNAL_SERVER))
            return
        }

        const creator = identities.get(userId)
        if (!creator) {
            writeResponse(req, res, newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT))
            return
        }

        const participantInfos = []
        for (const id of body.participantIds) {
            const identity = identities.get(id)
            if (!identity) {
                writeResponse(req, res, newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT))
                return
            }
            participantInfos.push({
                userId: id,
                username: identity.username,
                profilePicture: identity.profilePicture
            })
        }

        const result = await this.conversationService.createConversation(
            body.type,
            userId,
            creator.username,
            creator.profilePicture,
            participantInfos,
            body.name
        )

        writeResponse(req, res, result)
    }

    getConversations = async (req: Request, res: Response) => {
        const userId = req.userId!
        const page = Math.max(0, parseInt(req.query.page as string) || 0)
        const limit = Math.min(50, Math.max(1, parseInt(req.query.limit as string) || 20))

        const result = await this.conversationService.getConversations(userId, page, limit)
        writeResponse(req, res, result)
    }

    getConversation = async (req: Request, res: Response) => {
        const userId = req.userId!
        const conversationId = req.params.id

        const result = await this.conversationService.getConversation(conversationId, userId)
        writeResponse(req, res, result)
    }

    updateConversation = async (req: Request, res: Response) => {
        const userId = req.userId!
        const conversationId = req.params.id
        const body = req.body as UpdateConversationRequest

        const result = await this.conversationService.updateConversation(conversationId, userId, body)
        writeResponse(req, res, result)
    }

    leaveConversation = async (req: Request, res: Response) => {
        const userId = req.userId!
        const conversationId = req.params.id

        const result = await this.conversationService.leaveConversation(conversationId, userId)
        writeResponse(req, res, result)
    }

    addParticipant = async (req: Request, res: Response) => {
        const userId = req.userId!
        const conversationId = req.params.id
        const body = req.body as AddParticipantRequest

        if (!body.userId || !body.username) {
            writeResponse(req, res, newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT))
            return
        }

        let identities
        try {
            identities = await this.userServiceGateway.getIdentities([body.userId])
        } catch {
            writeResponse(req, res, newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_INTERNAL_SERVER))
            return
        }

        const identity = identities.get(body.userId)
        if (!identity) {
            writeResponse(req, res, newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_INVALID_INPUT))
            return
        }

        const result = await this.conversationService.addParticipant(conversationId, userId, {
            userId: body.userId,
            username: identity.username,
            profilePicture: identity.profilePicture
        })

        writeResponse(req, res, result)
    }

    removeParticipant = async (req: Request, res: Response) => {
        const userId = req.userId!
        const conversationId = req.params.id
        const targetUserId = req.params.userId

        const result = await this.conversationService.removeParticipant(conversationId, userId, targetUserId)
        writeResponse(req, res, result)
    }

    getUnreadCounts = async (req: Request, res: Response) => {
        const userId = req.userId!

        const result = await this.conversationService.getUnreadCounts(userId)
        writeResponse(req, res, result)
    }
}
