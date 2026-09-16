package vodcommentlike

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODCommentLikeRepo) DecrementLikeCount(ctx context.Context, commentId uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE vod_comments SET like_count = GREATEST(like_count - 1, 0) WHERE id = $1`,
		commentId,
	)
	if err != nil {
		logger.Errorf(ctx, "db exec error [decrementlikecount: %v]", err)
		return domains.ErrDatabaseIssue
	}
	return nil
}
