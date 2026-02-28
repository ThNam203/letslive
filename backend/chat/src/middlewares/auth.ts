import { NextFunction, Request, Response } from 'express'
import jwt from 'jsonwebtoken'
import logger from 'lib/logger'
import { RESPONSE_TEMPLATES, newResponseFromTemplate } from '../types/api-response'

interface JwtClaims {
    userId: string
    exp?: number
}

export function extractUserIdFromToken(token: string): string | null {
    try {
        // ParseUnverified — Kong already verified the signature
        const decoded = jwt.decode(token) as JwtClaims | null
        if (!decoded || !decoded.userId) return null
        if (decoded.exp && decoded.exp * 1000 < Date.now()) return null
        return decoded.userId
    } catch {
        return null
    }
}

export function extractUserIdFromCookie(cookieHeader: string | undefined): string | null {
    if (!cookieHeader) return null

    const cookies = cookieHeader.split(';').reduce(
        (acc, cookie) => {
            const [key, ...rest] = cookie.trim().split('=')
            acc[key] = rest.join('=')
            return acc
        },
        {} as Record<string, string>
    )

    const token = cookies['ACCESS_TOKEN']
    if (!token) return null
    return extractUserIdFromToken(token)
}

export function authMiddleware(req: Request, res: Response, next: NextFunction) {
    const userId = extractUserIdFromCookie(req.headers.cookie)
    if (!userId) {
        logger.debug('missing or invalid credentials in cookie')
        const resData = newResponseFromTemplate<void>(RESPONSE_TEMPLATES.RES_ERR_UNAUTHORIZED)
        resData.requestId = req.requestId ?? ''
        res.status(resData.statusCode).json(resData)
        return
    }

    req.userId = userId
    next()
}
