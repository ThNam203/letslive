package vodcommentlike

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODCommentLikeRepo) DeleteLike(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) error {
	cmdTag, err := r.db.Exec(ctx,
		`DELETE FROM vod_comment_likes WHERE comment_id = $1 AND user_id = $2`,
		commentId, userId,
	)
	if err != nil {
		logger.Errorf(ctx, "db exec error [deletelike: %v]", err)
		return domains.ErrDatabaseIssue
	}

	if cmdTag.RowsAffected() == 0 {
		return domains.ErrCommentNotLiked
	}

	return nil
}
