package livestream

import (
	"context"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresLivestreamRepo) Delete(ctx context.Context, livestreamId uuid.UUID) *response.Response[any] {
	result, err := r.dbConn.Exec(ctx, `
		DELETE FROM livestreams 
		WHERE id = $1
	`, livestreamId)
	if err != nil {
		logger.Errorf(ctx, "db exec error [deletelivestream id=%s: %v]", livestreamId, err)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	if result.RowsAffected() == 0 {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_LIVESTREAM_NOT_FOUND,
			nil,
			nil,
			nil,
		)
	}
	return nil
}
