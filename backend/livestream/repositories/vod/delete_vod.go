package vod

import (
	"context"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODRepo) Delete(ctx context.Context, id uuid.UUID) *response.Response[any] {
	result, err := r.dbConn.Exec(ctx, "delete from vods where id = $1", id)
	if err != nil {
		logger.Errorf(ctx, "db exec error [deletevod id=%s: %v]", id, err)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}
	if result.RowsAffected() == 0 {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_VOD_NOT_FOUND,
			nil,
			nil,
			nil,
		)
	}
	return nil
}
