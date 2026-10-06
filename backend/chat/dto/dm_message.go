package dto

import "sen1or/letslive/chat/domains"

type ReadReceiptResponse struct {
	UserID string `json:"userId"`
	ReadAt string `json:"readAt"`
}

type DmMessageResponse struct {
	ID             string                `json:"_id"`
	ConversationID string                `json:"conversationId"`
	SenderID       string                `json:"senderId"`
	SenderUsername string                `json:"senderUsername"`
	Type           domains.DmMessageType `json:"type"`
	Text           string                `json:"text"`
	ImageURLs      []string              `json:"imageUrls"`
	ReplyTo        *string               `json:"replyTo"`
	IsDeleted      bool                  `json:"isDeleted"`
	ReadBy         []ReadReceiptResponse `json:"readBy"`
	CreatedAt      string                `json:"createdAt"`
	UpdatedAt      string                `json:"updatedAt"`
	Version        int32                 `json:"__v"`
}

// SentDmMessageResponse is what POST /messages returns: the message plus the
// conversation's participant ids, as the Node service did.
type SentDmMessageResponse struct {
	DmMessageResponse
	ParticipantIDs []string `json:"participantIds"`
}

func FromDmMessage(m *domains.DmMessage) *DmMessageResponse {
	var replyTo *string
	if m.ReplyTo != nil {
		hex := m.ReplyTo.Hex()
		replyTo = &hex
	}

	imageURLs := m.ImageURLs
	if imageURLs == nil {
		imageURLs = []string{}
	}

	readBy := make([]ReadReceiptResponse, 0, len(m.ReadBy))
	for _, receipt := range m.ReadBy {
		readBy = append(readBy, ReadReceiptResponse{UserID: receipt.UserID, ReadAt: FormatTime(receipt.ReadAt)})
	}

	return &DmMessageResponse{
		ID:             m.ID.Hex(),
		ConversationID: m.ConversationID.Hex(),
		SenderID:       m.SenderID,
		SenderUsername: m.SenderUsername,
		Type:           m.Type,
		Text:           m.Text,
		ImageURLs:      imageURLs,
		ReplyTo:        replyTo,
		IsDeleted:      m.IsDeleted,
		ReadBy:         readBy,
		CreatedAt:      FormatTime(m.CreatedAt),
		UpdatedAt:      FormatTime(m.UpdatedAt),
		Version:        m.Version,
	}
}

func FromDmMessages(messages []domains.DmMessage) []DmMessageResponse {
	out := make([]DmMessageResponse, 0, len(messages))
	for i := range messages {
		out = append(out, *FromDmMessage(&messages[i]))
	}
	return out
}
