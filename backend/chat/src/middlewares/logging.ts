import { Request, Response, NextFunction } from 'express'

function loggingMiddleware(req: Request, res: Response, next: NextFunction): void {
  const startTime = Date.now()
  const originalWrite = res.write.bind(res)
  const originalEnd = res.end.bind(res)

  let responseBytes = 0

  res.write = ((chunk: any, ...args: any[]) => {
    if (chunk) responseBytes += Buffer.byteLength(chunk)
    return originalWrite(chunk, ...args)
  }) as typeof res.write

  res.end = ((chunk: any, ...args: any[]) => {
    if (chunk) responseBytes += Buffer.byteLength(chunk)
    return originalEnd(chunk, ...args)
  }) as typeof res.end

  res.on('finish', () => {
    const duration = Date.now() - startTime
    const ip =
      req.headers['x-forwarded-for']?.toString().split(',')[0].trim() ||
      req.socket?.remoteAddress ||
      'unknown address'

    const requestId = (req as any).requestId || req.headers['x-request-id'] || 'unknown'

    const fields = {
      requestId,
      duration,
      method: req.method,
      'remote#addr': ip,
      'response#bytes': responseBytes,
      'response#status': res.statusCode,
      uri: req.originalUrl,
    }

    const isHealthCheck = req.originalUrl === '/v1/health'
    const letsLiveError = res.getHeader('X-LetsLive-Error') as string | undefined

    if (res.statusCode >= 200 && res.statusCode < 300) {
      if (!isHealthCheck) {
        console.info('success api call', fields)
      }
    } else {
      if (letsLiveError) {
        console.error(`failed api call: ${letsLiveError}`, fields)
      } else {
        console.info('failed api call', fields)
      }
    }
  })

  next()
}

export default loggingMiddleware;
