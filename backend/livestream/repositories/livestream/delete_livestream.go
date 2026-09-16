package livestream

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresLivestreamRepo) Delete(ctx context.Context, livestreamId uuid.UUID) error {
	result, err := r.dbConn.Exec(ctx, `
		DELETE FROM livestreams 
		WHERE id = $1
	`, livestreamId)
	if err != nil {
		logger.Errorf(ctx, "db exec error [deletelivestream id=%s: %v]", livestreamId, err)
		return domains.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		return domains.ErrLivestreamNotFound
	}
	return nil
}
