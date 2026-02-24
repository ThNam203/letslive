package vodcomment

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresVODCommentRepo) GetByVODId(ctx context.Context, vodId uuid.UUID, page int, limit int) ([]domains.VODComment, *response.Response[any]) {
	offset := limit * page
	query := `
		SELECT id, vod_id, user_id, parent_id, content, is_deleted, like_count, reply_count, created_at, updated_at
		FROM vod_comments
		WHERE vod_id = $1 AND parent_id IS NULL
		ORDER BY created_at DESC
		OFFSET $2
		LIMIT $3
	`
	rows, err := r.db.Query(ctx, query, vodId, offset, limit)
	if err != nil {
		logger.Errorf(ctx, "db query error [getvodcommentsbyvodid: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	comments, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.VODComment])
	if err != nil {
		logger.Errorf(ctx, "db scan error [getvodcommentsbyvodid: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return comments, nil
}
