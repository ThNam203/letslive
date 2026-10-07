package dto

import "sen1or/letslive/chat/domains"

// ChatCommandResponse keeps the Node service's shape: "id", not "_id".
type ChatCommandResponse struct {
	ID          string                   `json:"id"`
	Scope       domains.ChatCommandScope `json:"scope"`
	OwnerID     string                   `json:"ownerId"`
	Name        string                   `json:"name"`
	Response    string                   `json:"response"`
	Description string                   `json:"description"`
	CreatedAt   string                   `json:"createdAt"`
}

type MyChatCommandsResponse struct {
	User    []ChatCommandResponse `json:"user"`
	Channel []ChatCommandResponse `json:"channel"`
}

func FromChatCommand(c *domains.ChatCommand) *ChatCommandResponse {
	return &ChatCommandResponse{
		ID:          c.ID.Hex(),
		Scope:       c.Scope,
		OwnerID:     c.OwnerID,
		Name:        c.Name,
		Response:    c.Response,
		Description: c.Description,
		CreatedAt:   FormatTime(c.CreatedAt),
	}
}

func FromChatCommands(commands []domains.ChatCommand) []ChatCommandResponse {
	out := make([]ChatCommandResponse, 0, len(commands))
	for i := range commands {
		out = append(out, *FromChatCommand(&commands[i]))
	}
	return out
}
