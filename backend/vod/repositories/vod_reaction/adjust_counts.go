package vodreaction

import (
	"context"
	"errors"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresVODReactionRepo) AdjustCounts(ctx context.Context, vodId uuid.UUID, likeDelta int64, dislikeDelta int64) (int64, error) {
	var likeCount int64
	err := r.db.QueryRow(ctx,
		`UPDATE vods
		SET like_count = GREATEST(like_count + $2, 0), dislike_count = GREATEST(dislike_count + $3, 0)
		WHERE id = $1
		RETURNING like_count`,
		vodId, likeDelta, dislikeDelta,
	).Scan(&likeCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domains.ErrVODNotFound
		}
		logger.Errorf(ctx, "db exec error [adjustvodreactioncounts: %v]", err)
		return 0, domains.ErrDatabaseIssue
	}
	return likeCount, nil
}
