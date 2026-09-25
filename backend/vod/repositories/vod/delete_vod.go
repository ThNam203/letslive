package vod

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.dbConn.Exec(ctx, "delete from vods where id = $1", id)
	if err != nil {
		logger.Errorf(ctx, "db exec error [deletevod id=%s: %v]", id, err)
		return domains.ErrDatabaseQuery
	}
	if result.RowsAffected() == 0 {
		return domains.ErrVODNotFound
	}
	return nil
}
