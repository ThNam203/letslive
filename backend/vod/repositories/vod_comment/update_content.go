package vodcomment

import (
	"context"
	"errors"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

// UpdateContent replaces the content of a comment that has not been deleted.
// A deleted comment reports ErrCommentNotFound: its content is not the
// author's to change any more.
func (r *postgresVODCommentRepo) UpdateContent(ctx context.Context, id uuid.UUID, content string) (*domains.VODComment, error) {
	query := `
		UPDATE vod_comments
		SET content = $2, is_edited = true, updated_at = now()
		WHERE id = $1 AND is_deleted = false
		RETURNING id, vod_id, user_id, parent_id, content, is_deleted, is_edited, like_count, reply_count, created_at, updated_at
	`
	rows, err := r.db.Query(ctx, query, id, content)
	if err != nil {
		logger.Errorf(ctx, "db query error [updatevodcommentcontent: %v]", err)
		return nil, domains.ErrCommentUpdateFailed
	}

	updatedComment, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.VODComment])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrCommentNotFound
		}
		logger.Errorf(ctx, "db scan error [updatevodcommentcontent: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}
	return &updatedComment, nil
}
