package vodcommentlike

import (
	"context"
	"errors"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *postgresVODCommentLikeRepo) InsertLike(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO vod_comment_likes (comment_id, user_id) VALUES ($1, $2)`,
		commentId, userId,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domains.ErrCommentAlreadyLiked
		}
		logger.Errorf(ctx, "db exec error [insertlike: %v]", err)
		return domains.ErrDatabaseIssue
	}
	return nil
}
