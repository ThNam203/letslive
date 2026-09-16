package vodcomment

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODCommentRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE vod_comments SET is_deleted = true, updated_at = now()
		WHERE id = $1
	`, id)
	if err != nil {
		logger.Errorf(ctx, "db exec error [softdeletevodcomment: %v]", err)
		return domains.ErrCommentDeleteFailed
	}
	return nil
}
