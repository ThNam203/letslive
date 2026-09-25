import { Request, Response } from 'express'
import { ConversationHandler } from '../handlers/conversationHandler'
import { ConversationService } from '../services/conversationService'
import { UserIdentity, UserServiceGateway } from '../gateway/userService'
import { RESPONSE_TEMPLATES, newResponseFromTemplate } from '../types/api-response'
import { ConversationType } from '../types/conversation'

const CREATOR_ID = '11111111-1111-1111-1111-111111111111'
const TARGET_ID = '22222222-2222-2222-2222-222222222222'

function identity(id: string, username: string): UserIdentity {
    return { id, username, profilePicture: null }
}

function makeGateway(identities: UserIdentity[]): UserServiceGateway {
    return {
        getIdentities: jest.fn(async () => new Map(identities.map((i) => [i.id, i])))
    } as unknown as UserServiceGateway
}

function makeService() {
    return {
        createConversation: jest.fn(async () => newResponseFromTemplate(RESPONSE_TEMPLATES.RES_SUCC_OK)),
        addParticipant: jest.fn(async () => newResponseFromTemplate(RESPONSE_TEMPLATES.RES_SUCC_OK))
    }
}

function makeReqRes(body: unknown, params: Record<string, string> = {}) {
    const req = { userId: CREATOR_ID, requestId: 'req-1', body, params } as unknown as Request
    const json = jest.fn()
    const status = jest.fn(() => ({ json }))
    const res = { status } as unknown as Response
    return { req, res, status, json }
}

describe('ConversationHandler: users without a username', () => {
    it('rejects creating a DM with a user who has not finished account setup', async () => {
        const service = makeService()
        const handler = new ConversationHandler(
            service as unknown as ConversationService,
            makeGateway([identity(CREATOR_ID, 'creator'), identity(TARGET_ID, '')])
        )
        const { req, res, status, json } = makeReqRes({ type: ConversationType.DM, participantIds: [TARGET_ID] })

        await handler.createConversation(req, res)

        expect(service.createConversation).not.toHaveBeenCalled()
        expect(status).toHaveBeenCalledWith(RESPONSE_TEMPLATES.RES_ERR_USER_SETUP_INCOMPLETE.statusCode)
        expect(json).toHaveBeenCalledWith(
            expect.objectContaining({ key: RESPONSE_TEMPLATES.RES_ERR_USER_SETUP_INCOMPLETE.key })
        )
    })

    it('rejects creating a conversation when the creator has not finished account setup', async () => {
        const service = makeService()
        const handler = new ConversationHandler(
            service as unknown as ConversationService,
            makeGateway([identity(CREATOR_ID, ''), identity(TARGET_ID, 'target')])
        )
        const { req, res, status } = makeReqRes({ type: ConversationType.DM, participantIds: [TARGET_ID] })

        await handler.createConversation(req, res)

        expect(service.createConversation).not.toHaveBeenCalled()
        expect(status).toHaveBeenCalledWith(RESPONSE_TEMPLATES.RES_ERR_USER_SETUP_INCOMPLETE.statusCode)
    })

    it('rejects adding a participant who has not finished account setup', async () => {
        const service = makeService()
        const handler = new ConversationHandler(
            service as unknown as ConversationService,
            makeGateway([identity(TARGET_ID, '')])
        )
        const { req, res, status } = makeReqRes(
            { userId: TARGET_ID, username: 'ignored' },
            { id: '507f1f77bcf86cd799439011' }
        )

        await handler.addParticipant(req, res)

        expect(service.addParticipant).not.toHaveBeenCalled()
        expect(status).toHaveBeenCalledWith(RESPONSE_TEMPLATES.RES_ERR_USER_SETUP_INCOMPLETE.statusCode)
    })

    it('creates the conversation when every user has a username', async () => {
        const service = makeService()
        const handler = new ConversationHandler(
            service as unknown as ConversationService,
            makeGateway([identity(CREATOR_ID, 'creator'), identity(TARGET_ID, 'target')])
        )
        const { req, res } = makeReqRes({ type: ConversationType.DM, participantIds: [TARGET_ID] })

        await handler.createConversation(req, res)

        expect(service.createConversation).toHaveBeenCalledWith(
            ConversationType.DM,
            CREATOR_ID,
            'creator',
            null,
            [{ userId: TARGET_ID, username: 'target', profilePicture: null }],
            undefined
        )
    })
})
