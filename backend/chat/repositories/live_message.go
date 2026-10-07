package repositories

import (
	"context"

	"sen1or/letslive/chat/domains"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// LiveMessagesCollection is the name Mongoose derives from the Message model.
const LiveMessagesCollection = "messages"

type LiveMessageRepository struct {
	coll *mongo.Collection
}

func NewLiveMessageRepository(db *mongo.Database) *LiveMessageRepository {
	return &LiveMessageRepository{coll: db.Collection(LiveMessagesCollection)}
}

func (r *LiveMessageRepository) Insert(ctx context.Context, message *domains.LiveMessage) error {
	if _, err := r.coll.InsertOne(ctx, message); err != nil {
		return dbError("insert live message", err)
	}
	return nil
}

// Latest returns the room's newest messages, newest first.
func (r *LiveMessageRepository) Latest(ctx context.Context, roomID string, limit int64) ([]domains.LiveMessage, error) {
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(limit)
	cursor, err := r.coll.Find(ctx, bson.M{"roomId": roomID}, opts)
	if err != nil {
		return nil, dbError("list live messages", err)
	}

	messages := []domains.LiveMessage{}
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, dbError("decode live messages", err)
	}
	return messages, nil
}
