package vodcommentlike

import (
	"context"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresVODCommentLikeRepo) GetUserLikedCommentIds(ctx context.Context, commentIds []uuid.UUID, userId uuid.UUID) ([]uuid.UUID, *response.Response[any]) {
	rows, err := r.db.Query(ctx, `
		SELECT comment_id FROM vod_comment_likes
		WHERE comment_id = ANY($1) AND user_id = $2
	`, commentIds, userId)
	if err != nil {
		logger.Errorf(ctx, "db query error [getuserlikedcommentids: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}
	defer rows.Close()

	likedIds, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		logger.Errorf(ctx, "db scan error [getuserlikedcommentids: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return likedIds, nil
}
