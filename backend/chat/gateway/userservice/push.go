package userservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type pushData struct {
	Type        string `json:"type"`
	ReferenceID string `json:"referenceId"`
	ActionURL   string `json:"actionUrl"`
}

type pushRequest struct {
	UserIDs []string `json:"userIds"`
	Title   string   `json:"title"`
	Body    string   `json:"body"`
	Data    pushData `json:"data"`
}

// SendDmPush asks the user service to push a direct message to the recipients'
// devices (POST /v1/internal/push). The user service skips users who are online
// and treats delivery as best-effort, so success only means "accepted".
// The deep link matches the web route /messages/[conversationId].
func (g *Gateway) SendDmPush(ctx context.Context, recipientIDs []string, senderName, text, conversationID string) error {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		return fmt.Errorf("resolve user service: %w", err)
	}

	body, err := json.Marshal(pushRequest{
		UserIDs: recipientIDs,
		Title:   senderName,
		Body:    text,
		Data: pushData{
			Type:        "dm",
			ReferenceID: conversationID,
			ActionURL:   "/messages/" + conversationID,
		},
	})
	if err != nil {
		return fmt.Errorf("encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/v1/internal/push", addr), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", requestID(ctx))

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("call user service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("user service returned %d", resp.StatusCode)
	}
	return nil
}
