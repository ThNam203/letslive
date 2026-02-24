package vodcomment

import (
	"context"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODCommentRepo) CountByVODId(ctx context.Context, vodId uuid.UUID) (int, *response.Response[any]) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM vod_comments WHERE vod_id = $1 AND parent_id IS NULL
	`, vodId).Scan(&count)
	if err != nil {
		logger.Errorf(ctx, "db query error [countvodcommentsbyvodid: %v]", err)
		return 0, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	return count, nil
}
