import mongoose from 'mongoose'

const Message = mongoose.model(
    'Message',
    new mongoose.Schema({
        roomId: {
            type: String,
            required: true,
            index: true
        },
        username: {
            type: String,
            required: true
        },
        userId: {
            type: String,
            required: true
        },
        text: {
            type: String
        },
        timestamp: { type: Date, default: Date.now }
    })
)

export { Message }
