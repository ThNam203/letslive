package domains

import (
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ChatCommand is a custom slash command, shaped like the Node service's
// Mongoose ChatCommand model (collection "chat_commands"). A channel command
// is owned by the channel's user id, which is also the live chat room id.
type ChatCommand struct {
	ID          bson.ObjectID    `bson:"_id"`
	Scope       ChatCommandScope `bson:"scope"`
	OwnerID     string           `bson:"ownerId"`
	Name        string           `bson:"name"`
	Response    string           `bson:"response"`
	Description string           `bson:"description"`
	CreatedAt   time.Time        `bson:"createdAt"`
	Version     int32            `bson:"__v"`
}

type ChatCommandScope string

const (
	ChatCommandScopeUser    ChatCommandScope = "user"
	ChatCommandScopeChannel ChatCommandScope = "channel"
)

func (s ChatCommandScope) Valid() bool {
	return s == ChatCommandScopeUser || s == ChatCommandScopeChannel
}

const (
	MaxChatCommandResponse    = 500
	MaxChatCommandDescription = 120
	MaxChatCommandsPerScope   = 50
	MaxChatCommandIDLength    = 36
)

var ChatCommandNamePattern = regexp.MustCompile(`^[a-z0-9_-]{1,32}$`)
