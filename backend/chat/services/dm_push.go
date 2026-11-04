package services

import (
	"context"
	"regexp"
	"slices"
	"time"

	"sen1or/letslive/chat/domains"
	"sen1or/letslive/shared/pkg/logger"
)

const (
	// ImagePushBody is the push body for a message that is only images. It is
	// the web i18n string messages.sent_an_image (web/lib/i18n/locales/en-US).
	ImagePushBody = "Sent an image"

	// MaxPushBodyLen is the user service limit for a push body, in characters.
	MaxPushBodyLen = 500

	// pushTimeout bounds one push call. Send does not wait for it.
	pushTimeout = 5 * time.Second
)

// imagePlaceholder matches the text the web client sends with image-only
// messages ("Sent an image" / "Sent 3 images"); see message-bubble.tsx.
var imagePlaceholder = regexp.MustCompile(`^Sent (an image|\d+ images)$`)

// PushClient delivers a direct-message push through the user service.
type PushClient interface {
	SendDmPush(ctx context.Context, recipientIDs []string, senderName, text, conversationID string) error
}

// dmPushBody is the notification text for message: a fixed string for
// image-only messages, otherwise the text cut to MaxPushBodyLen characters.
func dmPushBody(message *domains.DmMessage) string {
	if message.Type == domains.DmMessageTypeImage && len(message.ImageURLs) > 0 && imagePlaceholder.MatchString(message.Text) {
		return ImagePushBody
	}
	runes := []rune(message.Text)
	if len(runes) <= MaxPushBodyLen {
		return message.Text
	}
	return string(runes[:MaxPushBodyLen-1]) + "…"
}

// pushDm pushes message to every participant except the sender. Push is
// best-effort: a failure is logged and never reaches the caller.
func (s *DmMessageService) pushDm(ctx context.Context, participantIDs []string, senderID, senderName string, message *domains.DmMessage, conversationID string) {
	recipients := slices.DeleteFunc(slices.Clone(participantIDs), func(id string) bool { return id == senderID })
	if len(recipients) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, pushTimeout)
	defer cancel()
	if err := s.push.SendDmPush(ctx, recipients, senderName, dmPushBody(message), conversationID); err != nil {
		logger.Errorf(ctx, "failed to push dm %s: %v", message.ID.Hex(), err)
	}
}
