import mongoose from "mongoose";

const Message = mongoose.model(
    'Message',
    new mongoose.Schema({
        room: {
            type: String,
            required: true,
            index: true,
        },
        senderName: {
            type: String,
            required: true
        },
        senderId: {
            type: String,
            required: true
        },
        text: String,
        timestamp: { type: Date, default: Date.now },
    })
);

export { Message };