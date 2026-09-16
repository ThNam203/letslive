package vodcomment

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/jackc/pgx/v5"
)

func (r *postgresVODCommentRepo) Create(ctx context.Context, comment domains.VODComment) (*domains.VODComment, error) {
	query := `
		INSERT INTO vod_comments (vod_id, user_id, parent_id, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, vod_id, user_id, parent_id, content, is_deleted, like_count, reply_count, created_at, updated_at
	`
	rows, err := r.db.Query(ctx, query, comment.VODId, comment.UserId, comment.ParentId, comment.Content)
	if err != nil {
		logger.Errorf(ctx, "db query error [createvodcomment: %v]", err)
		return nil, domains.ErrCommentCreateFailed
	}

	createdComment, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.VODComment])
	if err != nil {
		logger.Errorf(ctx, "db scan error [createvodcomment: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}
	return &createdComment, nil
}
