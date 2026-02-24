import { NextFunction, Request, Response } from 'express'
import { v4 as uuidv4 } from 'uuid'

function requestIdMiddleware(req: Request, res: Response, next: NextFunction) {
    const existingId = req.header('X-Request-ID')
    const requestId = existingId || uuidv4()

    // Store it in the request object
    req.requestId = requestId

    // Add it to the response header
    res.setHeader('X-Request-ID', requestId)

    next()
}

export default requestIdMiddleware
