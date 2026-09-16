package vodcomment

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODCommentRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM vod_comments WHERE id = $1`, id)
	if err != nil {
		logger.Errorf(ctx, "db exec error [harddeletevodcomment: %v]", err)
		return domains.ErrCommentDeleteFailed
	}
	return nil
}
