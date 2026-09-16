package vodcomment

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODCommentRepo) DecrementReplyCount(ctx context.Context, commentId uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE vod_comments SET reply_count = GREATEST(reply_count - 1, 0) WHERE id = $1`,
		commentId,
	)
	if err != nil {
		logger.Errorf(ctx, "db exec error [decrementreplycount: %v]", err)
		return domains.ErrDatabaseIssue
	}
	return nil
}
