package services

import (
	"context"
	"fmt"
	"time"

	"sen1or/letslive/chat/domains"
	"sen1or/letslive/chat/gateway/userservice"
	"sen1or/letslive/chat/jsutil"
	"sen1or/letslive/chat/repositories"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ConversationService is a rule-by-rule port of the Node chat service's
// ConversationHandler + ConversationService. Checks run in the same order so
// the same request gets the same error.
type ConversationService struct {
	conversations *repositories.ConversationRepository
	messages      *repositories.DmMessageRepository
	users         *userservice.Gateway
}

func NewConversationService(
	conversations *repositories.ConversationRepository,
	messages *repositories.DmMessageRepository,
	users *userservice.Gateway,
) *ConversationService {
	return &ConversationService{conversations: conversations, messages: messages, users: users}
}

// OptionalString is a JSON field that may be absent, null or a string.
// Invalid is set when it was present with any other JSON type.
type OptionalString struct {
	Present bool
	Invalid bool
	Value   *string
}

type CreateConversationInput struct {
	Type           domains.ConversationType
	CreatorID      string
	ParticipantIDs []string
	Name           OptionalString
}

func parseObjectID(id string) (bson.ObjectID, bool) {
	objectID, err := bson.ObjectIDFromHex(id)
	return objectID, err == nil
}

// hasFinishedSetup: users who have not finished account setup have an empty
// username, which a participant record cannot hold.
func hasFinishedSetup(identity userservice.Identity) bool {
	return len(jsutil.Trim(identity.Username)) > 0
}

func validIdentity(identity userservice.Identity) bool {
	return jsutil.Length(identity.Username) <= domains.MaxUsernameLength &&
		(identity.ProfilePicture == nil || jsutil.Length(*identity.ProfilePicture) <= domains.MaxURLLength)
}

func dedupe(ids []string, exclude string) []string {
	out := make([]string, 0, len(ids))
	seen := map[string]struct{}{exclude: {}}
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// Create returns the conversation and whether it was newly created; an
// existing DM between the same two users is returned instead of a duplicate.
func (s *ConversationService) Create(ctx context.Context, input CreateConversationInput) (*domains.Conversation, bool, error) {
	// Node kept duplicate ids (and the creator, in a group) as extra
	// participants; they are dropped here. For a DM the creator stays so that
	// messaging yourself is still reported as such.
	exclude := ""
	if input.Type == domains.ConversationTypeGroup {
		exclude = input.CreatorID
	}
	participantIDs := dedupe(input.ParticipantIDs, exclude)

	if input.Type == domains.ConversationTypeGroup && len(participantIDs)+1 > domains.MaxGroupParticipants {
		return nil, false, domains.ErrTooManyParticipants
	}

	identities, err := s.users.GetIdentities(ctx, append([]string{input.CreatorID}, participantIDs...))
	if err != nil {
		return nil, false, err
	}

	creator, ok := identities[input.CreatorID]
	if !ok {
		return nil, false, domains.ErrInvalidInput
	}
	if !hasFinishedSetup(creator) {
		return nil, false, domains.ErrUserSetupIncomplete
	}

	participantIdentities := make([]userservice.Identity, 0, len(participantIDs))
	for _, id := range participantIDs {
		identity, ok := identities[id]
		if !ok {
			return nil, false, domains.ErrInvalidInput
		}
		if !hasFinishedSetup(identity) {
			return nil, false, domains.ErrUserSetupIncomplete
		}
		participantIdentities = append(participantIdentities, identity)
	}

	switch input.Type {
	case domains.ConversationTypeDM:
		if len(participantIDs) != 1 {
			return nil, false, domains.ErrInvalidInput
		}
		if participantIDs[0] == input.CreatorID {
			return nil, false, domains.ErrCannotMessageSelf
		}
		existing, err := s.conversations.FindDM(ctx, input.CreatorID, participantIDs[0])
		if err != nil {
			return nil, false, err
		}
		if existing != nil {
			return existing, false, nil
		}
	case domains.ConversationTypeGroup:
		if len(participantIDs) < 1 {
			return nil, false, domains.ErrInvalidInput
		}
	}

	// schema limits, which Mongoose checked when saving
	var name *string
	if input.Type == domains.ConversationTypeGroup {
		if input.Name.Invalid {
			return nil, false, domains.ErrInvalidInput
		}
		if input.Name.Value != nil && *input.Name.Value != "" {
			if jsutil.Length(*input.Name.Value) > domains.MaxConversationNameLen {
				return nil, false, domains.ErrInvalidInput
			}
			name = input.Name.Value
		}
	}
	if !validIdentity(creator) {
		return nil, false, domains.ErrInvalidInput
	}
	for _, identity := range participantIdentities {
		if !validIdentity(identity) {
			return nil, false, domains.ErrInvalidInput
		}
	}

	now := domains.Now()
	creatorRole := domains.ParticipantRoleMember
	if input.Type == domains.ConversationTypeGroup {
		creatorRole = domains.ParticipantRoleOwner
	}

	participants := []domains.Participant{newParticipant(input.CreatorID, creator, creatorRole, now)}
	for i, id := range participantIDs {
		participants = append(participants, newParticipant(id, participantIdentities[i], domains.ParticipantRoleMember, now))
	}

	conversation := &domains.Conversation{
		ID:           bson.NewObjectID(),
		Type:         input.Type,
		Name:         name,
		AvatarURL:    nil,
		CreatedBy:    input.CreatorID,
		Participants: participants,
		LastMessage:  nil,
		CreatedAt:    now,
		UpdatedAt:    now,
		Version:      0,
	}
	if err := s.conversations.Insert(ctx, conversation); err != nil {
		return nil, false, err
	}
	return conversation, true, nil
}

func newParticipant(userID string, identity userservice.Identity, role domains.ParticipantRole, joinedAt time.Time) domains.Participant {
	return domains.Participant{
		UserID:            userID,
		Username:          identity.Username,
		ProfilePicture:    identity.ProfilePicture,
		Role:              role,
		JoinedAt:          joinedAt,
		LastReadMessageID: nil,
		IsMuted:           false,
	}
}

func (s *ConversationService) List(ctx context.Context, userID string, page, limit int) ([]domains.Conversation, int64, error) {
	conversations, err := s.conversations.ListForUser(ctx, userID, int64(page*limit), int64(limit))
	if err != nil {
		return nil, 0, err
	}
	total, err := s.conversations.CountForUser(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	return conversations, total, nil
}

// load finds a conversation by its route id; a malformed id is reported as
// not found, like Node's ObjectId.isValid check.
func (s *ConversationService) load(ctx context.Context, id string) (*domains.Conversation, error) {
	objectID, ok := parseObjectID(id)
	if !ok {
		return nil, domains.ErrConversationNotFound
	}
	conversation, err := s.conversations.FindByID(ctx, objectID)
	if err != nil {
		return nil, err
	}
	if conversation == nil {
		return nil, domains.ErrConversationNotFound
	}
	return conversation, nil
}

func (s *ConversationService) Get(ctx context.Context, id, userID string) (*domains.Conversation, error) {
	conversation, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	if _, p := conversation.FindParticipant(userID); p == nil {
		return nil, domains.ErrNotParticipant
	}
	return conversation, nil
}

// loadGroupForManager runs the checks shared by update, add and remove:
// found, a group, the actor is a participant with owner or admin role.
func (s *ConversationService) loadGroupForManager(ctx context.Context, id, actorID string) (*domains.Conversation, error) {
	conversation, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	if conversation.Type != domains.ConversationTypeGroup {
		return nil, domains.ErrForbidden
	}
	_, actor := conversation.FindParticipant(actorID)
	if actor == nil {
		return nil, domains.ErrNotParticipant
	}
	if !actor.Role.CanManage() {
		return nil, domains.ErrInsufficientRole
	}
	return conversation, nil
}

type UpdateConversationInput struct {
	Name      OptionalString
	AvatarURL OptionalString
}

func sameString(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func (s *ConversationService) Update(ctx context.Context, id, userID string, input UpdateConversationInput) (*domains.Conversation, error) {
	conversation, err := s.loadGroupForManager(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if input.Name.Invalid || input.AvatarURL.Invalid {
		return nil, domains.ErrInvalidInput
	}
	if v := input.Name.Value; v != nil && jsutil.Length(*v) > domains.MaxConversationNameLen {
		return nil, domains.ErrInvalidInput
	}
	if v := input.AvatarURL.Value; v != nil && jsutil.Length(*v) > domains.MaxURLLength {
		return nil, domains.ErrInvalidInput
	}

	fields := bson.M{}
	if input.Name.Present && !sameString(input.Name.Value, conversation.Name) {
		fields["name"] = input.Name.Value
	}
	if input.AvatarURL.Present && !sameString(input.AvatarURL.Value, conversation.AvatarURL) {
		fields["avatarUrl"] = input.AvatarURL.Value
	}
	if len(fields) == 0 {
		return conversation, nil
	}

	updated, err := s.conversations.UpdateFields(ctx, conversation.ID, fields)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, domains.ErrConversationNotFound
	}
	return updated, nil
}

func (s *ConversationService) deleteWithMessages(ctx context.Context, id bson.ObjectID) error {
	if err := s.conversations.Delete(ctx, id); err != nil {
		return err
	}
	return s.messages.DeleteByConversation(ctx, id)
}

func (s *ConversationService) Leave(ctx context.Context, id, userID string) error {
	conversation, err := s.load(ctx, id)
	if err != nil {
		return err
	}
	idx, _ := conversation.FindParticipant(userID)
	if idx == -1 {
		return domains.ErrNotParticipant
	}

	if conversation.Type == domains.ConversationTypeDM {
		return s.deleteWithMessages(ctx, conversation.ID)
	}

	remaining := append(conversation.Participants[:idx:idx], conversation.Participants[idx+1:]...)
	if len(remaining) == 0 {
		return s.deleteWithMessages(ctx, conversation.ID)
	}

	// if the owner left, hand ownership to the first admin, or else the
	// longest-standing member
	hasOwner := false
	for _, p := range remaining {
		if p.Role == domains.ParticipantRoleOwner {
			hasOwner = true
			break
		}
	}
	if !hasOwner {
		newOwner := 0
		for i, p := range remaining {
			if p.Role == domains.ParticipantRoleAdmin {
				newOwner = i
				break
			}
		}
		remaining[newOwner].Role = domains.ParticipantRoleOwner
	}

	_, err = s.conversations.SetParticipants(ctx, conversation.ID, remaining)
	return err
}

func (s *ConversationService) AddParticipant(ctx context.Context, id, actorID, targetID string) (*domains.Conversation, error) {
	identities, err := s.users.GetIdentities(ctx, []string{targetID})
	if err != nil {
		return nil, err
	}
	identity, ok := identities[targetID]
	if !ok {
		return nil, domains.ErrInvalidInput
	}
	if !hasFinishedSetup(identity) {
		return nil, domains.ErrUserSetupIncomplete
	}

	conversation, err := s.loadGroupForManager(ctx, id, actorID)
	if err != nil {
		return nil, err
	}
	if _, existing := conversation.FindParticipant(targetID); existing != nil {
		return conversation, nil
	}
	if len(conversation.Participants) >= domains.MaxGroupParticipants {
		return nil, domains.ErrTooManyParticipants
	}
	if jsutil.Length(targetID) > domains.MaxUserIDLength || !validIdentity(identity) {
		return nil, domains.ErrInvalidInput
	}

	updated, err := s.conversations.AddParticipant(ctx, conversation.ID,
		newParticipant(targetID, identity, domains.ParticipantRoleMember, domains.Now()))
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, domains.ErrConversationNotFound
	}
	return updated, nil
}

func (s *ConversationService) RemoveParticipant(ctx context.Context, id, actorID, targetID string) (*domains.Conversation, error) {
	conversation, err := s.loadGroupForManager(ctx, id, actorID)
	if err != nil {
		return nil, err
	}
	idx, target := conversation.FindParticipant(targetID)
	if target == nil {
		return nil, domains.ErrNotParticipant
	}
	if target.Role == domains.ParticipantRoleOwner {
		return nil, domains.ErrInsufficientRole
	}

	remaining := append(conversation.Participants[:idx:idx], conversation.Participants[idx+1:]...)
	updated, err := s.conversations.SetParticipants(ctx, conversation.ID, remaining)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, domains.ErrConversationNotFound
	}
	return updated, nil
}

// UnreadCounts maps conversation id to the number of unread messages from
// other users; conversations with nothing unread are left out.
func (s *ConversationService) UnreadCounts(ctx context.Context, userID string) (map[string]int64, error) {
	conversations, err := s.conversations.ListForUser(ctx, userID, 0, 0)
	if err != nil {
		return nil, err
	}

	counts := map[string]int64{}
	for i := range conversations {
		_, participant := conversations[i].FindParticipant(userID)
		if participant == nil {
			continue
		}
		count, err := s.messages.CountUnread(ctx, conversations[i].ID, participant.LastReadMessageID, userID)
		if err != nil {
			return nil, fmt.Errorf("unread count for %s: %w", conversations[i].ID.Hex(), err)
		}
		if count > 0 {
			counts[conversations[i].ID.Hex()] = count
		}
	}
	return counts, nil
}
