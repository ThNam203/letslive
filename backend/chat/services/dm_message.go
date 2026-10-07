package services

import (
	"context"
	"slices"

	"sen1or/letslive/chat/domains"
	"sen1or/letslive/chat/dto"
	"sen1or/letslive/chat/events"
	"sen1or/letslive/chat/jsutil"
	"sen1or/letslive/chat/repositories"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// DmMessageService is a rule-by-rule port of the Node chat service's
// DmMessageService, plus the pushes the Node DM socket used to send.
type DmMessageService struct {
	conversations *repositories.ConversationRepository
	messages      *repositories.DmMessageRepository
	notifier      *events.Notifier
}

func NewDmMessageService(
	conversations *repositories.ConversationRepository,
	messages *repositories.DmMessageRepository,
	notifier *events.Notifier,
) *DmMessageService {
	return &DmMessageService{conversations: conversations, messages: messages, notifier: notifier}
}

// loadAsParticipant finds the conversation and the caller's participant record.
func (s *DmMessageService) loadAsParticipant(ctx context.Context, conversationID, userID string) (*domains.Conversation, *domains.Participant, error) {
	objectID, ok := parseObjectID(conversationID)
	if !ok {
		return nil, nil, domains.ErrConversationNotFound
	}
	conversation, err := s.conversations.FindByID(ctx, objectID)
	if err != nil {
		return nil, nil, err
	}
	if conversation == nil {
		return nil, nil, domains.ErrConversationNotFound
	}
	_, participant := conversation.FindParticipant(userID)
	if participant == nil {
		return nil, nil, domains.ErrNotParticipant
	}
	return conversation, participant, nil
}

// List returns a page of messages oldest first. before is ignored unless it
// is a valid message id.
func (s *DmMessageService) List(ctx context.Context, conversationID, userID, before string, limit int) ([]domains.DmMessage, error) {
	conversation, _, err := s.loadAsParticipant(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}

	var beforeID *bson.ObjectID
	if id, ok := parseObjectID(before); ok {
		beforeID = &id
	}

	messages, err := s.messages.FindPage(ctx, conversation.ID, beforeID, int64(min(max(limit, 1), 100)))
	if err != nil {
		return nil, err
	}
	slices.Reverse(messages)
	return messages, nil
}

type SendDmMessageInput struct {
	ConversationID string
	SenderID       string
	Text           string
	Type           domains.DmMessageType
	ImageURLs      []string
	ReplyTo        string
}

// Send saves the message, updates the conversation preview and pushes
// dm:new_message to every participant, the sender included so their other
// tabs see it too.
func (s *DmMessageService) Send(ctx context.Context, input SendDmMessageInput) (*domains.DmMessage, []string, error) {
	conversation, participant, err := s.loadAsParticipant(ctx, input.ConversationID, input.SenderID)
	if err != nil {
		return nil, nil, err
	}

	// the display name comes from the participant record, never the payload
	senderUsername := participant.Username

	trimmed := jsutil.Trim(input.Text)
	if len(trimmed) == 0 || jsutil.Length(input.Text) > domains.MaxDmMessageLength {
		return nil, nil, domains.ErrInvalidInput
	}

	imageURLs := []string{}
	if input.Type == domains.DmMessageTypeImage && input.ImageURLs != nil {
		imageURLs = input.ImageURLs
	}

	var replyTo *bson.ObjectID
	if id, ok := parseObjectID(input.ReplyTo); ok {
		replyTo = &id
	}

	now := domains.Now()
	message := &domains.DmMessage{
		ID:             bson.NewObjectID(),
		ConversationID: conversation.ID,
		SenderID:       input.SenderID,
		SenderUsername: senderUsername,
		Type:           input.Type,
		Text:           trimmed,
		ImageURLs:      imageURLs,
		ReplyTo:        replyTo,
		IsDeleted:      false,
		ReadBy:         []domains.ReadReceipt{{UserID: input.SenderID, ReadAt: now}},
		CreatedAt:      now,
		UpdatedAt:      now,
		Version:        0,
	}
	if err := s.messages.Insert(ctx, message); err != nil {
		return nil, nil, err
	}

	lastMessage := domains.LastMessage{
		ID:             message.ID,
		SenderID:       input.SenderID,
		SenderUsername: senderUsername,
		Text:           jsutil.Truncate(trimmed, domains.LastMessagePreviewLen),
		CreatedAt:      message.CreatedAt,
	}
	if err := s.conversations.SetLastMessage(ctx, conversation.ID, lastMessage); err != nil {
		return nil, nil, err
	}

	participantIDs := conversation.ParticipantIDs()
	s.notifier.ToUsers(ctx, participantIDs, events.DmNewMessage, events.NewMessage{
		ConversationID: conversation.ID.Hex(),
		Message:        dto.FromDmMessage(message),
	})

	return message, participantIDs, nil
}

func parseMessageIDs(conversationID, messageID string) (bson.ObjectID, bson.ObjectID, bool) {
	conversationObjectID, ok := parseObjectID(conversationID)
	if !ok {
		return bson.ObjectID{}, bson.ObjectID{}, false
	}
	messageObjectID, ok := parseObjectID(messageID)
	if !ok {
		return bson.ObjectID{}, bson.ObjectID{}, false
	}
	return conversationObjectID, messageObjectID, true
}

func (s *DmMessageService) Edit(ctx context.Context, conversationID, messageID, userID, text string) (*domains.DmMessage, error) {
	conversationObjectID, messageObjectID, ok := parseMessageIDs(conversationID, messageID)
	if !ok {
		return nil, domains.ErrDmMessageNotFound
	}

	trimmed := jsutil.Trim(text)
	if len(trimmed) == 0 || jsutil.Length(text) > domains.MaxDmMessageLength {
		return nil, domains.ErrInvalidInput
	}

	message, err := s.messages.FindInConversation(ctx, messageObjectID, conversationObjectID)
	if err != nil {
		return nil, err
	}
	if message == nil {
		return nil, domains.ErrDmMessageNotFound
	}
	if message.SenderID != userID {
		return nil, domains.ErrForbidden
	}
	if message.IsDeleted {
		return nil, domains.ErrDmMessageNotFound
	}
	if message.Text == trimmed {
		return message, nil
	}

	updated, err := s.messages.UpdateFields(ctx, message.ID, bson.M{"text": trimmed})
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, domains.ErrDmMessageNotFound
	}
	return updated, nil
}

// Delete soft-deletes the sender's own message. (In Node this always failed:
// the schema's required validator rejected the emptied text.)
func (s *DmMessageService) Delete(ctx context.Context, conversationID, messageID, userID string) error {
	conversationObjectID, messageObjectID, ok := parseMessageIDs(conversationID, messageID)
	if !ok {
		return domains.ErrDmMessageNotFound
	}

	message, err := s.messages.FindInConversation(ctx, messageObjectID, conversationObjectID)
	if err != nil {
		return err
	}
	if message == nil {
		return domains.ErrDmMessageNotFound
	}
	if message.SenderID != userID {
		return domains.ErrForbidden
	}
	if message.IsDeleted && message.Text == "" {
		return nil
	}

	_, err = s.messages.UpdateFields(ctx, message.ID, bson.M{"isDeleted": true, "text": ""})
	return err
}

// MarkAsRead moves the caller's read marker to messageID, or to the latest
// message when messageID is empty or not a valid id, and pushes
// dm:read_receipt to the other participants.
func (s *DmMessageService) MarkAsRead(ctx context.Context, conversationID, userID string, messageID *string) error {
	conversation, _, err := s.loadAsParticipant(ctx, conversationID, userID)
	if err != nil {
		return err
	}

	var readUpTo *bson.ObjectID
	if messageID != nil {
		if id, ok := parseObjectID(*messageID); ok {
			readUpTo = &id
		}
	}
	if readUpTo == nil {
		latest, err := s.messages.FindLatest(ctx, conversation.ID)
		if err != nil {
			return err
		}
		if latest != nil {
			readUpTo = &latest.ID
		}
	}

	if readUpTo != nil {
		if err := s.conversations.SetLastRead(ctx, conversation.ID, userID, *readUpTo); err != nil {
			return err
		}
	}

	var receiptMessageID *string
	if messageID != nil && *messageID != "" {
		receiptMessageID = messageID
	}
	s.notifier.ToUsers(ctx, conversation.OtherParticipantIDs(userID), events.DmReadReceipt, events.ReadReceipt{
		ConversationID: conversation.ID.Hex(),
		UserID:         userID,
		MessageID:      receiptMessageID,
		ReadAt:         dto.FormatTime(domains.Now()),
	})
	return nil
}

// Typing pushes a typing start or stop to the other participants. Nothing is
// stored.
func (s *DmMessageService) Typing(ctx context.Context, conversationID, userID string, started bool) error {
	conversation, participant, err := s.loadAsParticipant(ctx, conversationID, userID)
	if err != nil {
		return err
	}

	eventType := events.DmUserStoppedTyping
	if started {
		eventType = events.DmUserTyping
	}
	s.notifier.ToUsers(ctx, conversation.OtherParticipantIDs(userID), eventType, events.Typing{
		ConversationID: conversation.ID.Hex(),
		UserID:         userID,
		Username:       participant.Username,
	})
	return nil
}
