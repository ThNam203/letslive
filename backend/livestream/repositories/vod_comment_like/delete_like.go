package vodcommentlike

import (
	"context"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODCommentLikeRepo) DeleteLike(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) *response.Response[any] {
	cmdTag, err := r.db.Exec(ctx,
		`DELETE FROM vod_comment_likes WHERE comment_id = $1 AND user_id = $2`,
		commentId, userId,
	)
	if err != nil {
		logger.Errorf(ctx, "db exec error [deletelike: %v]", err)
		return response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_ISSUE, nil, nil, nil)
	}

	if cmdTag.RowsAffected() == 0 {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_VOD_COMMENT_NOT_LIKED, nil, nil, nil,
		)
	}

	return nil
}
