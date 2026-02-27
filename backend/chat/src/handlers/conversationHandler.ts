import { Request, Response } from 'express'
import { ConversationService } from '../services/conversationService'
import { RESPONSE_TEMPLATES, newResponseFromTemplate, Response as ServiceResponse } from '../types/api-response'
import { CreateConversationRequest, UpdateConversationRequest, AddParticipantRequest } from '../types/conversation'

function writeResponse(req: Request, res: Response, resData: ServiceResponse<any>) {
    resData.requestId = req.requestId ?? ''
    res.status(resData.statusCode).json(resData)
}

export class ConversationHandler {
    constructor(private conversationService: ConversationService) {}

    createConversation = async (req: Request, res: Response) => {
        const userId = req.userId!
        const body = req.body as CreateConversationRequest

        if (!body.type || !['dm', 'group'].includes(body.type)) {
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

        // For now, we only have userId info from the token. Participant usernames
        // will need to be provided by the client or fetched from user service.
        // Using minimal info here — the frontend sends participant info.
        const participantInfos = body.participantIds.map((id) => ({
            userId: id,
            username: (req.body.participantUsernames?.[id] as string) || id,
            displayName: (req.body.participantDisplayNames?.[id] as string) || null,
            profilePicture: (req.body.participantProfilePictures?.[id] as string) || null
        }))

        const result = await this.conversationService.createConversation(
            body.type,
            userId,
            req.body.creatorUsername || userId,
            req.body.creatorDisplayName || null,
            req.body.creatorProfilePicture || null,
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

        const result = await this.conversationService.addParticipant(conversationId, userId, {
            userId: body.userId,
            username: body.username,
            displayName: body.displayName || null,
            profilePicture: body.profilePicture || null
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
