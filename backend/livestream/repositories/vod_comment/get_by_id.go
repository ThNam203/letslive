package vodcomment

import (
	"context"
	"errors"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresVODCommentRepo) GetById(ctx context.Context, id uuid.UUID) (*domains.VODComment, *response.Response[any]) {
	query := `
		SELECT id, vod_id, user_id, parent_id, content, is_deleted, like_count, reply_count, created_at, updated_at
		FROM vod_comments
		WHERE id = $1
	`
	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		logger.Errorf(ctx, "db query error [getvodcommentbyid: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	comment, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.VODComment])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_VOD_COMMENT_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}
		logger.Errorf(ctx, "db scan error [getvodcommentbyid: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return &comment, nil
}
