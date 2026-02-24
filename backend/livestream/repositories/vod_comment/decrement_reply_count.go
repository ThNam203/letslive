package vodcomment

import (
	"context"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODCommentRepo) DecrementReplyCount(ctx context.Context, commentId uuid.UUID) *response.Response[any] {
	_, err := r.db.Exec(ctx,
		`UPDATE vod_comments SET reply_count = GREATEST(reply_count - 1, 0) WHERE id = $1`,
		commentId,
	)
	if err != nil {
		logger.Errorf(ctx, "db exec error [decrementreplycount: %v]", err)
		return response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_ISSUE, nil, nil, nil)
	}
	return nil
}
