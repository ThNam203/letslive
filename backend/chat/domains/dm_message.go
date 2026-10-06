package domains

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Documents in the dmmessages collection, shaped like the Node service's
// Mongoose DmMessage model.

type DmMessageType string

const (
	DmMessageTypeText   DmMessageType = "text"
	DmMessageTypeImage  DmMessageType = "image"
	DmMessageTypeSystem DmMessageType = "system"
)

const (
	MaxDmMessageLength    = 2000
	LastMessagePreviewLen = 100
)

func (t DmMessageType) Valid() bool {
	return t == DmMessageTypeText || t == DmMessageTypeImage || t == DmMessageTypeSystem
}

type ReadReceipt struct {
	UserID string    `bson:"userId"`
	ReadAt time.Time `bson:"readAt"`
}

type DmMessage struct {
	ID             bson.ObjectID  `bson:"_id"`
	ConversationID bson.ObjectID  `bson:"conversationId"`
	SenderID       string         `bson:"senderId"`
	SenderUsername string         `bson:"senderUsername"`
	Type           DmMessageType  `bson:"type"`
	Text           string         `bson:"text"`
	ImageURLs      []string       `bson:"imageUrls"`
	ReplyTo        *bson.ObjectID `bson:"replyTo"`
	IsDeleted      bool           `bson:"isDeleted"`
	ReadBy         []ReadReceipt  `bson:"readBy"`
	CreatedAt      time.Time      `bson:"createdAt"`
	UpdatedAt      time.Time      `bson:"updatedAt"`
	Version        int32          `bson:"__v"`
}
