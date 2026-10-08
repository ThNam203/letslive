package domains

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

type VODReactionType string

const (
	VODReactionLike    VODReactionType = "like"
	VODReactionDislike VODReactionType = "dislike"
)

type VODReactionRepository interface {
	WithTx(tx pgx.Tx) VODReactionRepository
	// GetReaction returns nil when the user has not reacted to the VOD.
	GetReaction(ctx context.Context, vodId uuid.UUID, userId uuid.UUID) (*VODReactionType, error)
	// GetReactionForUpdate is GetReaction that also locks the row until the tx ends.
	GetReactionForUpdate(ctx context.Context, vodId uuid.UUID, userId uuid.UUID) (*VODReactionType, error)
	// InsertReaction reports false when the user already has a reaction on the VOD.
	InsertReaction(ctx context.Context, vodId uuid.UUID, userId uuid.UUID, reaction VODReactionType) (bool, error)
	UpdateReaction(ctx context.Context, vodId uuid.UUID, userId uuid.UUID, reaction VODReactionType) error
	// DeleteReaction returns the removed reaction, or nil when there was none.
	DeleteReaction(ctx context.Context, vodId uuid.UUID, userId uuid.UUID) (*VODReactionType, error)
	// AdjustCounts adds the deltas to the VOD's like and dislike counters and returns the new like count.
	AdjustCounts(ctx context.Context, vodId uuid.UUID, likeDelta int64, dislikeDelta int64) (int64, error)
}
