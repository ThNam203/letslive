import pino from 'pino'

// ensures consistent log format across all services
// maybe a standalone package?
export default pino({
    level: process.env.LOG_LEVEL || 'info',
    formatters: {
        level: (label) => {
            return { level: label }
        }
    },
    timestamp: pino.stdTimeFunctions.isoTime
})
