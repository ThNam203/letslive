package events

import (
	"github.com/gofrs/uuid/v5"
)

// Notification event types.
const (
	NotificationRequested = "notification.requested"
)

// NotificationRequestedEvent is emitted when a service wants to send a notification to a user.
// This allows any service to trigger notifications without direct HTTP calls to the user service.
type NotificationRequestedEvent struct {
	UserId      uuid.UUID  `json:"userId"`
	Type        string     `json:"type"`
	Title       string     `json:"title"`
	Message     string     `json:"message"`
	ActionUrl   *string    `json:"actionUrl,omitempty"`
	ActionLabel *string    `json:"actionLabel,omitempty"`
	ReferenceId *uuid.UUID `json:"referenceId,omitempty"`
}
