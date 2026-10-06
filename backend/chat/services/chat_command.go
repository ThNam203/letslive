package services

import (
	"context"
	"strings"

	"sen1or/letslive/chat/domains"
	"sen1or/letslive/chat/jsutil"
	"sen1or/letslive/chat/repositories"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ChatCommandService is a rule-by-rule port of the Node chatCommandService.
// Every validation failure is invalid_input, as it was there.
type ChatCommandService struct {
	commands *repositories.ChatCommandRepository
}

func NewChatCommandService(commands *repositories.ChatCommandRepository) *ChatCommandService {
	return &ChatCommandService{commands: commands}
}

func normalizeCommandName(name string) (string, bool) {
	normalized := strings.ToLower(jsutil.Trim(name))
	return normalized, domains.ChatCommandNamePattern.MatchString(normalized)
}

func validCommandResponse(response string) bool {
	n := jsutil.Length(response)
	return n > 0 && n <= domains.MaxChatCommandResponse
}

func validCommandDescription(description string) bool {
	return jsutil.Length(description) <= domains.MaxChatCommandDescription
}

// ListForRoom returns the room's channel commands plus, for a signed-in
// viewer, their own user commands. A room id is its channel owner's user id.
func (s *ChatCommandService) ListForRoom(ctx context.Context, roomID, viewerID string) ([]domains.ChatCommand, error) {
	owners := map[domains.ChatCommandScope]string{domains.ChatCommandScopeChannel: roomID}
	if viewerID != "" {
		owners[domains.ChatCommandScopeUser] = viewerID
	}
	return s.commands.ListByOwners(ctx, owners)
}

// ListMine splits the user's commands by scope, each sorted by name.
func (s *ChatCommandService) ListMine(ctx context.Context, userID string) (user, channel []domains.ChatCommand, err error) {
	commands, err := s.commands.ListByOwners(ctx, map[domains.ChatCommandScope]string{
		domains.ChatCommandScopeUser:    userID,
		domains.ChatCommandScopeChannel: userID,
	})
	if err != nil {
		return nil, nil, err
	}

	user, channel = []domains.ChatCommand{}, []domains.ChatCommand{}
	for _, command := range commands {
		if command.Scope == domains.ChatCommandScopeUser {
			user = append(user, command)
		} else {
			channel = append(channel, command)
		}
	}
	return user, channel, nil
}

type CreateChatCommandInput struct {
	Scope       domains.ChatCommandScope
	Name        string
	Response    string
	Description string
}

func (s *ChatCommandService) Create(ctx context.Context, userID string, input CreateChatCommandInput) (*domains.ChatCommand, error) {
	name, ok := normalizeCommandName(input.Name)
	if !ok || !validCommandResponse(input.Response) || !validCommandDescription(input.Description) || !input.Scope.Valid() {
		return nil, domains.ErrInvalidInput
	}

	count, err := s.commands.CountByOwner(ctx, input.Scope, userID)
	if err != nil {
		return nil, err
	}
	if count >= domains.MaxChatCommandsPerScope {
		return nil, domains.ErrInvalidInput
	}

	command := &domains.ChatCommand{
		ID:          bson.NewObjectID(),
		Scope:       input.Scope,
		OwnerID:     userID,
		Name:        name,
		Response:    input.Response,
		Description: input.Description,
		CreatedAt:   domains.Now(),
		Version:     0,
	}
	if err := s.commands.Insert(ctx, command); err != nil {
		return nil, err
	}
	return command, nil
}

// loadOwned finds a command and checks the caller owns it. A malformed id is
// treated as a missing command.
func (s *ChatCommandService) loadOwned(ctx context.Context, userID, id string) (*domains.ChatCommand, error) {
	objectID, ok := parseObjectID(id)
	if !ok {
		return nil, nil
	}
	command, err := s.commands.FindByID(ctx, objectID)
	if err != nil || command == nil {
		return nil, err
	}
	if command.OwnerID != userID {
		return nil, domains.ErrForbidden
	}
	return command, nil
}

type UpdateChatCommandInput struct {
	Name        OptionalString
	Response    OptionalString
	Description OptionalString
}

// valueOf is a present field's string; null and wrongly typed values fail.
func valueOf(field OptionalString) (string, bool) {
	if field.Invalid || field.Value == nil {
		return "", false
	}
	return *field.Value, true
}

func (s *ChatCommandService) Update(ctx context.Context, userID, id string, input UpdateChatCommandInput) (*domains.ChatCommand, error) {
	command, err := s.loadOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if command == nil {
		return nil, domains.ErrInvalidInput
	}

	fields := bson.M{}
	if input.Name.Present {
		raw, _ := valueOf(input.Name) // a null name is "", which the pattern rejects
		name, ok := normalizeCommandName(raw)
		if !ok {
			return nil, domains.ErrInvalidInput
		}
		if name != command.Name {
			fields["name"], command.Name = name, name
		}
	}
	if input.Response.Present {
		response, ok := valueOf(input.Response)
		if !ok || !validCommandResponse(response) {
			return nil, domains.ErrInvalidInput
		}
		if response != command.Response {
			fields["response"], command.Response = response, response
		}
	}
	if input.Description.Present {
		description, ok := valueOf(input.Description)
		if !ok || !validCommandDescription(description) {
			return nil, domains.ErrInvalidInput
		}
		if description != command.Description {
			fields["description"], command.Description = description, description
		}
	}

	if len(fields) > 0 {
		if err := s.commands.UpdateFields(ctx, command.ID, fields); err != nil {
			return nil, err
		}
	}
	return command, nil
}

// Delete is idempotent: a missing command is already gone.
func (s *ChatCommandService) Delete(ctx context.Context, userID, id string) error {
	command, err := s.loadOwned(ctx, userID, id)
	if err != nil || command == nil {
		return err
	}
	return s.commands.Delete(ctx, command.ID)
}
