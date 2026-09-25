package vodcomment

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODCommentRepo) CountByVODId(ctx context.Context, vodId uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM vod_comments WHERE vod_id = $1 AND parent_id IS NULL
	`, vodId).Scan(&count)
	if err != nil {
		logger.Errorf(ctx, "db query error [countvodcommentsbyvodid: %v]", err)
		return 0, domains.ErrDatabaseQuery
	}

	return count, nil
}
