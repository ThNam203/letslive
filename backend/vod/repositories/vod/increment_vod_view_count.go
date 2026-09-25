package vod

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODRepo) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	query := `
        update vods
        set view_count = view_count + 1
        where id = $1
    `
	result, err := r.dbConn.Exec(ctx, query, id)
	if err != nil {
		logger.Errorf(ctx, "db exec error [incrementvodviewcount id=%s: %v]", id, err)
		return domains.ErrVODUpdateFailed
	}

	if result.RowsAffected() == 0 {
		logger.Warnf(ctx, "attempted to increment view count for non-existent vod id %s", id)
		return domains.ErrVODNotFound
	}

	return nil
}
