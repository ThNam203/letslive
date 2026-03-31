import mongoose from 'mongoose'

const Message = mongoose.model(
    'Message',
    new mongoose.Schema({
        roomId: {
            type: String,
            required: true,
            maxlength: 36,
            index: true
        },
        username: {
            type: String,
            required: true,
            maxlength: 50
        },
        userId: {
            type: String,
            required: true,
            maxlength: 36
        },
        text: {
            type: String,
            required: true,
            maxlength: 500
        },
        timestamp: { type: Date, default: Date.now }
    })
)

export { Message }
