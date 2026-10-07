package repositories

import (
	"context"
	"errors"

	"sen1or/letslive/chat/domains"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// DmMessagesCollection is the name Mongoose derives from the DmMessage model.
const DmMessagesCollection = "dmmessages"

type DmMessageRepository struct {
	coll *mongo.Collection
}

func NewDmMessageRepository(db *mongo.Database) *DmMessageRepository {
	return &DmMessageRepository{coll: db.Collection(DmMessagesCollection)}
}

func (r *DmMessageRepository) Insert(ctx context.Context, message *domains.DmMessage) error {
	if _, err := r.coll.InsertOne(ctx, message); err != nil {
		return dbError("insert dm message", err)
	}
	return nil
}

// FindPage returns up to limit messages newest first; before, when set,
// restricts the page to messages with a smaller id.
func (r *DmMessageRepository) FindPage(ctx context.Context, conversationID bson.ObjectID, before *bson.ObjectID, limit int64) ([]domains.DmMessage, error) {
	filter := bson.M{"conversationId": conversationID}
	if before != nil {
		filter["_id"] = bson.M{"$lt": *before}
	}
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(limit)

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, dbError("list dm messages", err)
	}

	messages := []domains.DmMessage{}
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, dbError("decode dm messages", err)
	}
	return messages, nil
}

// FindInConversation returns nil, nil when the message does not exist in that
// conversation.
func (r *DmMessageRepository) FindInConversation(ctx context.Context, id, conversationID bson.ObjectID) (*domains.DmMessage, error) {
	return r.findOne(ctx, bson.M{"_id": id, "conversationId": conversationID}, nil, "find dm message")
}

// FindLatest returns the newest message of the conversation, or nil, nil.
func (r *DmMessageRepository) FindLatest(ctx context.Context, conversationID bson.ObjectID) (*domains.DmMessage, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	return r.findOne(ctx, bson.M{"conversationId": conversationID}, opts, "find latest dm message")
}

func (r *DmMessageRepository) findOne(ctx context.Context, filter bson.M, opts *options.FindOneOptionsBuilder, op string) (*domains.DmMessage, error) {
	var message domains.DmMessage
	var err error
	if opts != nil {
		err = r.coll.FindOne(ctx, filter, opts).Decode(&message)
	} else {
		err = r.coll.FindOne(ctx, filter).Decode(&message)
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, dbError(op, err)
	}
	return &message, nil
}

// UpdateFields sets fields and bumps updatedAt, returning the updated message.
func (r *DmMessageRepository) UpdateFields(ctx context.Context, id bson.ObjectID, fields bson.M) (*domains.DmMessage, error) {
	set := bson.M{"updatedAt": domains.Now()}
	for key, value := range fields {
		set[key] = value
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var message domains.DmMessage
	err := r.coll.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": set}, opts).Decode(&message)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, dbError("update dm message", err)
	}
	return &message, nil
}

// CountUnread counts live messages from other users, newer than afterID when
// the reader has a read marker.
func (r *DmMessageRepository) CountUnread(ctx context.Context, conversationID bson.ObjectID, afterID *bson.ObjectID, readerID string) (int64, error) {
	filter := bson.M{
		"conversationId": conversationID,
		"senderId":       bson.M{"$ne": readerID},
		"isDeleted":      false,
	}
	if afterID != nil {
		filter["_id"] = bson.M{"$gt": *afterID}
	}
	count, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, dbError("count unread", err)
	}
	return count, nil
}

func (r *DmMessageRepository) DeleteByConversation(ctx context.Context, conversationID bson.ObjectID) error {
	if _, err := r.coll.DeleteMany(ctx, bson.M{"conversationId": conversationID}); err != nil {
		return dbError("delete dm messages", err)
	}
	return nil
}
