package vodcommentlike

import (
	"context"
	"errors"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *postgresVODCommentLikeRepo) InsertLike(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) *response.Response[any] {
	_, err := r.db.Exec(ctx,
		`INSERT INTO vod_comment_likes (comment_id, user_id) VALUES ($1, $2)`,
		commentId, userId,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return response.NewResponseFromTemplate[any](
				response.RES_ERR_VOD_COMMENT_ALREADY_LIKED, nil, nil, nil,
			)
		}
		logger.Errorf(ctx, "db exec error [insertlike: %v]", err)
		return response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_ISSUE, nil, nil, nil)
	}
	return nil
}
