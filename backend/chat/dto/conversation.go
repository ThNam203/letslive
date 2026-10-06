package dto

import (
	"time"

	"sen1or/letslive/chat/domains"
)

// The response shapes reproduce Mongoose's toObject() output as Express
// serialises it: _id and __v included, ObjectIds as hex, dates as
// JSON.stringify(Date) prints them, and nullable fields as null.

const jsDateLayout = "2006-01-02T15:04:05.000Z"

func FormatTime(t time.Time) string {
	return t.UTC().Format(jsDateLayout)
}

type ParticipantResponse struct {
	UserID            string                  `json:"userId"`
	Username          string                  `json:"username"`
	ProfilePicture    *string                 `json:"profilePicture"`
	Role              domains.ParticipantRole `json:"role"`
	JoinedAt          string                  `json:"joinedAt"`
	LastReadMessageID *string                 `json:"lastReadMessageId"`
	IsMuted           bool                    `json:"isMuted"`
}

type LastMessageResponse struct {
	ID             string `json:"_id"`
	SenderID       string `json:"senderId"`
	SenderUsername string `json:"senderUsername"`
	Text           string `json:"text"`
	CreatedAt      string `json:"createdAt"`
}

type ConversationResponse struct {
	ID           string                   `json:"_id"`
	Type         domains.ConversationType `json:"type"`
	Name         *string                  `json:"name"`
	AvatarURL    *string                  `json:"avatarUrl"`
	CreatedBy    string                   `json:"createdBy"`
	Participants []ParticipantResponse    `json:"participants"`
	LastMessage  *LastMessageResponse     `json:"lastMessage"`
	CreatedAt    string                   `json:"createdAt"`
	UpdatedAt    string                   `json:"updatedAt"`
	Version      int32                    `json:"__v"`
}

func FromConversation(c *domains.Conversation) *ConversationResponse {
	participants := make([]ParticipantResponse, 0, len(c.Participants))
	for _, p := range c.Participants {
		var lastRead *string
		if p.LastReadMessageID != nil {
			hex := p.LastReadMessageID.Hex()
			lastRead = &hex
		}
		participants = append(participants, ParticipantResponse{
			UserID:            p.UserID,
			Username:          p.Username,
			ProfilePicture:    p.ProfilePicture,
			Role:              p.Role,
			JoinedAt:          FormatTime(p.JoinedAt),
			LastReadMessageID: lastRead,
			IsMuted:           p.IsMuted,
		})
	}

	var lastMessage *LastMessageResponse
	if c.LastMessage != nil {
		lastMessage = &LastMessageResponse{
			ID:             c.LastMessage.ID.Hex(),
			SenderID:       c.LastMessage.SenderID,
			SenderUsername: c.LastMessage.SenderUsername,
			Text:           c.LastMessage.Text,
			CreatedAt:      FormatTime(c.LastMessage.CreatedAt),
		}
	}

	return &ConversationResponse{
		ID:           c.ID.Hex(),
		Type:         c.Type,
		Name:         c.Name,
		AvatarURL:    c.AvatarURL,
		CreatedBy:    c.CreatedBy,
		Participants: participants,
		LastMessage:  lastMessage,
		CreatedAt:    FormatTime(c.CreatedAt),
		UpdatedAt:    FormatTime(c.UpdatedAt),
		Version:      c.Version,
	}
}

func FromConversations(conversations []domains.Conversation) []ConversationResponse {
	out := make([]ConversationResponse, 0, len(conversations))
	for i := range conversations {
		out = append(out, *FromConversation(&conversations[i]))
	}
	return out
}
