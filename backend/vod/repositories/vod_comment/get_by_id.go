package vodcomment

import (
	"context"
	"errors"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresVODCommentRepo) GetById(ctx context.Context, id uuid.UUID) (*domains.VODComment, error) {
	query := `
		SELECT id, vod_id, user_id, parent_id, content, is_deleted, like_count, reply_count, created_at, updated_at
		FROM vod_comments
		WHERE id = $1
	`
	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		logger.Errorf(ctx, "db query error [getvodcommentbyid: %v]", err)
		return nil, domains.ErrDatabaseQuery
	}

	comment, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.VODComment])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrCommentNotFound
		}
		logger.Errorf(ctx, "db scan error [getvodcommentbyid: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}
	return &comment, nil
}
