package vod

import (
	"context"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODRepo) IncrementViewCount(ctx context.Context, id uuid.UUID) *response.Response[any] {
	query := `
        update vods
        set view_count = view_count + 1
        where id = $1
    `
	result, err := r.dbConn.Exec(ctx, query, id)
	if err != nil {
		logger.Errorf(ctx, "db exec error [incrementvodviewcount id=%s: %v]", id, err)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_VOD_UPDATE_FAILED,
			nil,
			nil,
			nil,
		)
	}

	if result.RowsAffected() == 0 {
		logger.Warnf(ctx, "attempted to increment view count for non-existent vod id %s", id)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_VOD_NOT_FOUND,
			nil,
			nil,
			nil,
		)
	}

	return nil
}
