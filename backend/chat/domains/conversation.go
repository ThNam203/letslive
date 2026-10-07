package domains

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Documents in the conversations collection. Field names, BSON types and
// defaults match what the Node service's Mongoose schema writes, because both
// services share the collection while the migration is in progress.

type ConversationType string

const (
	ConversationTypeDM    ConversationType = "dm"
	ConversationTypeGroup ConversationType = "group"
)

type ParticipantRole string

const (
	ParticipantRoleOwner  ParticipantRole = "owner"
	ParticipantRoleAdmin  ParticipantRole = "admin"
	ParticipantRoleMember ParticipantRole = "member"
)

// MaxGroupParticipants counts the owner. Creating a group looks every member
// up in one user service batch call, which accepts at most 100 ids, so this
// cannot go higher without batching that call.
const MaxGroupParticipants = 100

// schema limits the Mongoose models enforced on save
const (
	MaxUserIDLength        = 36
	MaxUsernameLength      = 50
	MaxConversationNameLen = 100
	MaxURLLength           = 2048
)

type Participant struct {
	UserID            string          `bson:"userId"`
	Username          string          `bson:"username"`
	ProfilePicture    *string         `bson:"profilePicture"`
	Role              ParticipantRole `bson:"role"`
	JoinedAt          time.Time       `bson:"joinedAt"`
	LastReadMessageID *bson.ObjectID  `bson:"lastReadMessageId"`
	IsMuted           bool            `bson:"isMuted"`
}

type LastMessage struct {
	ID             bson.ObjectID `bson:"_id"`
	SenderID       string        `bson:"senderId"`
	SenderUsername string        `bson:"senderUsername"`
	Text           string        `bson:"text"`
	CreatedAt      time.Time     `bson:"createdAt"`
}

type Conversation struct {
	ID           bson.ObjectID    `bson:"_id"`
	Type         ConversationType `bson:"type"`
	Name         *string          `bson:"name"`
	AvatarURL    *string          `bson:"avatarUrl"`
	CreatedBy    string           `bson:"createdBy"`
	Participants []Participant    `bson:"participants"`
	LastMessage  *LastMessage     `bson:"lastMessage"`
	CreatedAt    time.Time        `bson:"createdAt"`
	UpdatedAt    time.Time        `bson:"updatedAt"`
	Version      int32            `bson:"__v"`
}

func (c *Conversation) FindParticipant(userID string) (int, *Participant) {
	for i := range c.Participants {
		if c.Participants[i].UserID == userID {
			return i, &c.Participants[i]
		}
	}
	return -1, nil
}

func (c *Conversation) ParticipantIDs() []string {
	ids := make([]string, 0, len(c.Participants))
	for _, p := range c.Participants {
		ids = append(ids, p.UserID)
	}
	return ids
}

func (c *Conversation) OtherParticipantIDs(userID string) []string {
	ids := make([]string, 0, len(c.Participants))
	for _, p := range c.Participants {
		if p.UserID != userID {
			ids = append(ids, p.UserID)
		}
	}
	return ids
}

func (r ParticipantRole) CanManage() bool {
	return r == ParticipantRoleOwner || r == ParticipantRoleAdmin
}

// Now is the timestamp every write uses. Mongo stores milliseconds, so
// truncating keeps the value returned to the client equal to the stored one.
func Now() time.Time {
	return time.Now().UTC().Truncate(time.Millisecond)
}
