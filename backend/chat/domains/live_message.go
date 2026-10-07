package domains

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// LiveMessage is a live chat line, shaped like the Node service's Mongoose
// Message model (collection "messages").
type LiveMessage struct {
	ID             bson.ObjectID `bson:"_id"`
	RoomID         string        `bson:"roomId"`
	Username       string        `bson:"username"`
	UserID         string        `bson:"userId"`
	Text           string        `bson:"text"`
	ProfilePicture *string       `bson:"profilePicture"`
	Timestamp      time.Time     `bson:"timestamp"`
	Version        int32         `bson:"__v"`
}

const (
	MaxLiveMessageLength = 500
	LiveHistorySize      = 50
	MaxRoomIDLength      = 36
)
