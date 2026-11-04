// Package push delivers mobile push notifications through FCM.
//
// Reusable entry point: build a Message (FromNotification maps a stored
// notification into one, other features build their own) and call
// MessageSender.SendMessage. Delivery is skipped for users with a live realtime
// connection and removes device tokens that FCM reports as dead.
package push

import (
	"context"
	"errors"
	"fmt"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
)

// PushSender sends a push for a stored notification.
type PushSender interface {
	Send(ctx context.Context, userID uuid.UUID, n domains.Notification) error
}

// PresenceChecker reports whether the user has an active realtime connection.
type PresenceChecker interface {
	IsOnline(ctx context.Context, userID uuid.UUID) (bool, error)
}

// MessageSender is the payload-agnostic entry point shared by notification
// pushes and other pushes (for example direct messages).
type MessageSender interface {
	SendMessage(ctx context.Context, userID uuid.UUID, msg Message) error
}

// Message is the transport-neutral push payload. Data values must be strings
// (FCM requirement); empty values are omitted by the builders.
type Message struct {
	Title string
	Body  string
	Data  map[string]string
}

// Data keys of notification pushes. The mobile deep-link handler relies on them.
const (
	DataKeyType        = "type"
	DataKeyReferenceID = "referenceId"
	DataKeyActionURL   = "actionUrl"
)

// FromNotification maps a stored notification into a push Message. The title
// and message are sent as stored (the notification is not localized).
func FromNotification(n domains.Notification) Message {
	data := map[string]string{}
	if n.Type != "" {
		data[DataKeyType] = n.Type
	}
	if n.ReferenceId != nil {
		data[DataKeyReferenceID] = n.ReferenceId.String()
	}
	if n.ActionUrl != nil && *n.ActionUrl != "" {
		data[DataKeyActionURL] = *n.ActionUrl
	}
	return Message{Title: n.Title, Body: n.Message, Data: data}
}

// TokenStore is the slice of the device-token repository the sender needs.
type TokenStore interface {
	ListTokensByUserID(ctx context.Context, userID uuid.UUID) ([]domains.DeviceToken, error)
	DeleteByToken(ctx context.Context, token string) error
}

// TokenStatus is the per-token outcome of a gateway send.
type TokenStatus int

const (
	StatusSent TokenStatus = iota
	// StatusUnregistered: FCM says the token is no longer valid.
	StatusUnregistered
	// StatusInvalidArgument: FCM rejected the request for this token. It can
	// be the token or the message, so it is only trusted as a bad token when
	// another token of the same send succeeded.
	StatusInvalidArgument
	StatusFailed
)

type TokenResult struct {
	Token  string
	Status TokenStatus
	Err    error
}

// Gateway is the FCM boundary. A returned error means the whole call failed.
type Gateway interface {
	SendToTokens(ctx context.Context, tokens []string, msg Message) ([]TokenResult, error)
}

// Service sends pushes synchronously. Wrap it in Async for fire-and-forget use.
type Service struct {
	tokens   TokenStore
	presence PresenceChecker
	gateway  Gateway
}

func NewService(tokens TokenStore, presence PresenceChecker, gateway Gateway) *Service {
	return &Service{tokens: tokens, presence: presence, gateway: gateway}
}

func (s *Service) Send(ctx context.Context, userID uuid.UUID, n domains.Notification) error {
	return s.SendMessage(ctx, userID, FromNotification(n))
}

func (s *Service) SendMessage(ctx context.Context, userID uuid.UUID, msg Message) error {
	online, err := s.presence.IsOnline(ctx, userID)
	if err != nil {
		// a duplicate notification beats a lost one
		logger.Warnf(ctx, "presence unavailable for user %s, sending push anyway: %v", userID, err)
	} else if online {
		logger.Debugf(ctx, "user %s has a live realtime connection, skipping push", userID)
		return nil
	}

	rows, err := s.tokens.ListTokensByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("list device tokens: %w", err)
	}
	if len(rows) == 0 {
		return nil
	}

	tokens := make([]string, len(rows))
	for i, r := range rows {
		tokens[i] = r.Token
	}

	results, err := s.gateway.SendToTokens(ctx, tokens, msg)
	if err != nil {
		return fmt.Errorf("fcm send: %w", err)
	}
	return s.cleanup(ctx, userID, results)
}

// cleanup deletes dead tokens. Errors never include the token itself.
func (s *Service) cleanup(ctx context.Context, userID uuid.UUID, results []TokenResult) error {
	anySent := false
	for _, r := range results {
		if r.Status == StatusSent {
			anySent = true
			break
		}
	}

	var errs []error
	failed := 0
	for _, r := range results {
		switch {
		case r.Status == StatusUnregistered, r.Status == StatusInvalidArgument && anySent:
			if err := s.tokens.DeleteByToken(ctx, r.Token); err != nil {
				errs = append(errs, fmt.Errorf("delete dead token: %w", err))
			}
		case r.Status != StatusSent:
			failed++
		}
	}
	if failed > 0 {
		errs = append(errs, fmt.Errorf("push to %d of %d device(s) of user %s failed", failed, len(results), userID))
	}
	return errors.Join(errs...)
}

// Disabled is the sender used when no FCM credentials are configured.
type Disabled struct{}

func NewDisabled() Disabled { return Disabled{} }

func (Disabled) Send(ctx context.Context, userID uuid.UUID, _ domains.Notification) error {
	logger.Debugf(ctx, "push disabled, not sending to user %s", userID)
	return nil
}

func (Disabled) SendMessage(ctx context.Context, userID uuid.UUID, _ Message) error {
	logger.Debugf(ctx, "push disabled, not sending to user %s", userID)
	return nil
}
