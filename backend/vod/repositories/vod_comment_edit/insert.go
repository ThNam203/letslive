package vodcommentedit

import (
	"context"
	"errors"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// uniqueViolationCode is the PostgreSQL SQLSTATE for a unique constraint
// violation.
const uniqueViolationCode = "23505"

// Insert appends previousContent to the comment's history as the next
// version. Two concurrent edits compute the same next version, so the unique
// constraint on (comment_id, version) makes the loser fail here instead of
// silently sharing a version number with the winner.
func (r *postgresVODCommentEditRepo) Insert(ctx context.Context, commentId uuid.UUID, previousContent string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO vod_comment_edits (comment_id, version, previous_content)
		SELECT $1, COALESCE(MAX(version), 0) + 1, $2
		FROM vod_comment_edits
		WHERE comment_id = $1
	`, commentId, previousContent)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			logger.Warnf(ctx, "concurrent edit on comment %s rejected", commentId)
			return domains.ErrCommentUpdateFailed
		}
		logger.Errorf(ctx, "db exec error [insertvodcommentedit: %v]", err)
		return domains.ErrCommentUpdateFailed
	}
	return nil
}
