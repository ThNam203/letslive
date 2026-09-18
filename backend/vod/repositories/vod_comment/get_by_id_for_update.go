package vodcomment

import (
	"context"
	"errors"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

// GetByIdForUpdate reads a comment and holds a row lock until the surrounding
// transaction ends. Editing reads the comment to record what it said before
// the edit, so that read has to be serialised against a concurrent edit of
// the same comment — otherwise both would record the same "before" text and
// one of the two versions would be lost.
func (r *postgresVODCommentRepo) GetByIdForUpdate(ctx context.Context, id uuid.UUID) (*domains.VODComment, error) {
	query := `
		SELECT id, vod_id, user_id, parent_id, content, is_deleted, is_edited, like_count, reply_count, created_at, updated_at
		FROM vod_comments
		WHERE id = $1
		FOR UPDATE
	`
	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		logger.Errorf(ctx, "db query error [getvodcommentbyidforupdate: %v]", err)
		return nil, domains.ErrDatabaseQuery
	}

	comment, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.VODComment])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrCommentNotFound
		}
		logger.Errorf(ctx, "db scan error [getvodcommentbyidforupdate: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}
	return &comment, nil
}
