package vodcomment

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/jackc/pgx/v5"
)

func (r *postgresVODCommentRepo) Create(ctx context.Context, comment domains.VODComment) (*domains.VODComment, *response.Response[any]) {
	query := `
		INSERT INTO vod_comments (vod_id, user_id, parent_id, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, vod_id, user_id, parent_id, content, is_deleted, like_count, reply_count, created_at, updated_at
	`
	rows, err := r.db.Query(ctx, query, comment.VODId, comment.UserId, comment.ParentId, comment.Content)
	if err != nil {
		logger.Errorf(ctx, "db query error [createvodcomment: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_VOD_COMMENT_CREATE_FAILED,
			nil,
			nil,
			nil,
		)
	}

	createdComment, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.VODComment])
	if err != nil {
		logger.Errorf(ctx, "db scan error [createvodcomment: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return &createdComment, nil
}
