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

// ConversationsCollection is the name Mongoose derives from the Conversation model.
const ConversationsCollection = "conversations"

type ConversationRepository struct {
	coll *mongo.Collection
}

func NewConversationRepository(db *mongo.Database) *ConversationRepository {
	return &ConversationRepository{coll: db.Collection(ConversationsCollection)}
}

func dbError(op string, err error) error {
	return fmt.Errorf("%w: %s: %v", domains.ErrDatabaseIssue, op, err)
}

// FindByID returns nil, nil when no conversation has this id.
func (r *ConversationRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domains.Conversation, error) {
	return r.findOne(ctx, bson.M{"_id": id}, "find conversation")
}

// FindDM returns the DM between exactly these two users, or nil, nil.
func (r *ConversationRepository) FindDM(ctx context.Context, userA, userB string) (*domains.Conversation, error) {
	filter := bson.M{
		"type":                domains.ConversationTypeDM,
		"participants.userId": bson.M{"$all": bson.A{userA, userB}},
		"$expr":               bson.M{"$eq": bson.A{bson.M{"$size": "$participants"}, 2}},
	}
	return r.findOne(ctx, filter, "find dm")
}

func (r *ConversationRepository) findOne(ctx context.Context, filter bson.M, op string) (*domains.Conversation, error) {
	var conversation domains.Conversation
	err := r.coll.FindOne(ctx, filter).Decode(&conversation)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, dbError(op, err)
	}
	return &conversation, nil
}

// ListForUser returns the user's conversations, most recently updated first.
// A limit of 0 means no limit.
func (r *ConversationRepository) ListForUser(ctx context.Context, userID string, skip, limit int64) ([]domains.Conversation, error) {
	opts := options.Find().SetSort(bson.D{{Key: "updatedAt", Value: -1}}).SetSkip(skip)
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.coll.Find(ctx, bson.M{"participants.userId": userID}, opts)
	if err != nil {
		return nil, dbError("list conversations", err)
	}

	conversations := []domains.Conversation{}
	if err := cursor.All(ctx, &conversations); err != nil {
		return nil, dbError("decode conversations", err)
	}
	return conversations, nil
}

func (r *ConversationRepository) CountForUser(ctx context.Context, userID string) (int64, error) {
	count, err := r.coll.CountDocuments(ctx, bson.M{"participants.userId": userID})
	if err != nil {
		return 0, dbError("count conversations", err)
	}
	return count, nil
}

func (r *ConversationRepository) Insert(ctx context.Context, conversation *domains.Conversation) error {
	if _, err := r.coll.InsertOne(ctx, conversation); err != nil {
		return dbError("insert conversation", err)
	}
	return nil
}

func (r *ConversationRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	if _, err := r.coll.DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		return dbError("delete conversation", err)
	}
	return nil
}

// UpdateFields sets plain fields and bumps updatedAt, returning the updated
// document (nil if it no longer exists).
func (r *ConversationRepository) UpdateFields(ctx context.Context, id bson.ObjectID, fields bson.M) (*domains.Conversation, error) {
	set := bson.M{"updatedAt": domains.Now()}
	for key, value := range fields {
		set[key] = value
	}
	return r.updateOne(ctx, bson.M{"_id": id}, bson.M{"$set": set}, "update conversation")
}

// AddParticipant appends a participant. Like Mongoose, any change to the
// participants array increments __v.
func (r *ConversationRepository) AddParticipant(ctx context.Context, id bson.ObjectID, participant domains.Participant) (*domains.Conversation, error) {
	update := bson.M{
		"$push": bson.M{"participants": participant},
		"$set":  bson.M{"updatedAt": domains.Now()},
		"$inc":  bson.M{"__v": 1},
	}
	return r.updateOne(ctx, bson.M{"_id": id}, update, "add participant")
}

// SetParticipants replaces the participants array (removal, leave, ownership
// transfer), incrementing __v as Mongoose does for array rewrites.
func (r *ConversationRepository) SetParticipants(ctx context.Context, id bson.ObjectID, participants []domains.Participant) (*domains.Conversation, error) {
	update := bson.M{
		"$set": bson.M{"participants": participants, "updatedAt": domains.Now()},
		"$inc": bson.M{"__v": 1},
	}
	return r.updateOne(ctx, bson.M{"_id": id}, update, "set participants")
}

func (r *ConversationRepository) SetLastMessage(ctx context.Context, id bson.ObjectID, lastMessage domains.LastMessage) error {
	update := bson.M{"$set": bson.M{"lastMessage": lastMessage, "updatedAt": domains.Now()}}
	if _, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, update); err != nil {
		return dbError("set last message", err)
	}
	return nil
}

// SetLastRead moves the user's read marker. It writes nothing when the marker
// is already there, so updatedAt only moves on a real change, as with
// Mongoose's dirty tracking.
func (r *ConversationRepository) SetLastRead(ctx context.Context, id bson.ObjectID, userID string, messageID bson.ObjectID) error {
	filter := bson.M{
		"_id": id,
		"participants": bson.M{"$elemMatch": bson.M{
			"userId":            userID,
			"lastReadMessageId": bson.M{"$ne": messageID},
		}},
	}
	update := bson.M{"$set": bson.M{
		"participants.$.lastReadMessageId": messageID,
		"updatedAt":                        domains.Now(),
	}}
	if _, err := r.coll.UpdateOne(ctx, filter, update); err != nil {
		return dbError("set last read", err)
	}
	return nil
}

func (r *ConversationRepository) updateOne(ctx context.Context, filter, update bson.M, op string) (*domains.Conversation, error) {
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var conversation domains.Conversation
	err := r.coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&conversation)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, dbError(op, err)
	}
	return &conversation, nil
}
