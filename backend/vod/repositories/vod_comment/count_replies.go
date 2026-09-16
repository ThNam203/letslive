package vodcomment

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODCommentRepo) CountReplies(ctx context.Context, parentId uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM vod_comments WHERE parent_id = $1
	`, parentId).Scan(&count)
	if err != nil {
		logger.Errorf(ctx, "db query error [countvodcommentreplies: %v]", err)
		return 0, domains.ErrDatabaseQuery
	}

	return count, nil
}
