package repositories

import (
	"context"
	"errors"
	"fmt"

	"sen1or/letslive/chat/domains"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ChatCommandsCollection is set explicitly on the Mongoose model.
const ChatCommandsCollection = "chat_commands"

type ChatCommandRepository struct {
	coll *mongo.Collection
}

func NewChatCommandRepository(db *mongo.Database) *ChatCommandRepository {
	return &ChatCommandRepository{coll: db.Collection(ChatCommandsCollection)}
}

// writeError reports the unique (scope, ownerId, name) index as
// ErrAlreadyExists so the service can tell a name clash from a failure.
func writeError(op string, err error) error {
	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("%w: %s", domains.ErrAlreadyExists, op)
	}
	return dbError(op, err)
}

// ListByOwners returns the commands matching any (scope, ownerId) pair,
// sorted by name.
func (r *ChatCommandRepository) ListByOwners(ctx context.Context, owners map[domains.ChatCommandScope]string) ([]domains.ChatCommand, error) {
	filters := bson.A{}
	for scope, ownerID := range owners {
		filters = append(filters, bson.M{"scope": scope, "ownerId": ownerID})
	}

	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	cursor, err := r.coll.Find(ctx, bson.M{"$or": filters}, opts)
	if err != nil {
		return nil, dbError("list chat commands", err)
	}

	commands := []domains.ChatCommand{}
	if err := cursor.All(ctx, &commands); err != nil {
		return nil, dbError("decode chat commands", err)
	}
	return commands, nil
}

func (r *ChatCommandRepository) CountByOwner(ctx context.Context, scope domains.ChatCommandScope, ownerID string) (int64, error) {
	count, err := r.coll.CountDocuments(ctx, bson.M{"scope": scope, "ownerId": ownerID})
	if err != nil {
		return 0, dbError("count chat commands", err)
	}
	return count, nil
}

// FindByID returns nil, nil when no command has this id.
func (r *ChatCommandRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domains.ChatCommand, error) {
	var command domains.ChatCommand
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&command)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, dbError("find chat command", err)
	}
	return &command, nil
}

func (r *ChatCommandRepository) Insert(ctx context.Context, command *domains.ChatCommand) error {
	if _, err := r.coll.InsertOne(ctx, command); err != nil {
		return writeError("insert chat command", err)
	}
	return nil
}

func (r *ChatCommandRepository) UpdateFields(ctx context.Context, id bson.ObjectID, fields bson.M) error {
	if _, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": fields}); err != nil {
		return writeError("update chat command", err)
	}
	return nil
}

func (r *ChatCommandRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	if _, err := r.coll.DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		return dbError("delete chat command", err)
	}
	return nil
}
