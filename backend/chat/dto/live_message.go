package dto

import "sen1or/letslive/chat/domains"

// LiveMessageResponse is a history item, shaped like the Mongoose Message doc.
type LiveMessageResponse struct {
	ID             string  `json:"_id"`
	RoomID         string  `json:"roomId"`
	Username       string  `json:"username"`
	UserID         string  `json:"userId"`
	Text           string  `json:"text"`
	ProfilePicture *string `json:"profilePicture"`
	Timestamp      string  `json:"timestamp"`
	Version        int32   `json:"__v"`
}

func FromLiveMessage(m *domains.LiveMessage) *LiveMessageResponse {
	return &LiveMessageResponse{
		ID:             m.ID.Hex(),
		RoomID:         m.RoomID,
		Username:       m.Username,
		UserID:         m.UserID,
		Text:           m.Text,
		ProfilePicture: m.ProfilePicture,
		Timestamp:      FormatTime(m.Timestamp),
		Version:        m.Version,
	}
}

func FromLiveMessages(messages []domains.LiveMessage) []LiveMessageResponse {
	out := make([]LiveMessageResponse, 0, len(messages))
	for i := range messages {
		out = append(out, *FromLiveMessage(&messages[i]))
	}
	return out
}
