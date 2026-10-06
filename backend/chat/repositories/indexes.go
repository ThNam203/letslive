package repositories

import (
	"context"

	"sen1or/letslive/shared/pkg/logger"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// EnsureIndexes creates the same indexes the Mongoose schemas declare. With
// identical keys the default names match too, so on a database Node already
// indexed this is a no-op. A failure is logged rather than fatal: an existing
// index with different options must not keep the service from starting.
func EnsureIndexes(ctx context.Context, db *mongo.Database) {
	indexes := map[string][]mongo.IndexModel{
		ConversationsCollection: {
			{Keys: bson.D{{Key: "participants.userId", Value: 1}, {Key: "updatedAt", Value: -1}}},
			{Keys: bson.D{{Key: "type", Value: 1}, {Key: "participants.userId", Value: 1}}},
		},
		DmMessagesCollection: {
			{Keys: bson.D{{Key: "conversationId", Value: 1}}},
			{Keys: bson.D{{Key: "conversationId", Value: 1}, {Key: "createdAt", Value: -1}}},
		},
	}

	for collection, models := range indexes {
		if _, err := db.Collection(collection).Indexes().CreateMany(ctx, models); err != nil {
			logger.Errorf(ctx, "failed to ensure indexes on %s: %v", collection, err)
		}
	}
}
