package devicerepo

import (
	"context"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresDeviceRepo) DeleteByToken(ctx context.Context, token string) error {
	if _, err := r.dbConn.Exec(ctx, `DELETE FROM device_tokens WHERE token = $1`, token); err != nil {
		logger.Errorf(ctx, "failed to delete device token: %s", err)
		return domains.ErrDatabaseQuery
	}
	return nil
}

func (r *postgresDeviceRepo) DeleteByUserAndToken(ctx context.Context, userID uuid.UUID, token string) error {
	if _, err := r.dbConn.Exec(ctx, `DELETE FROM device_tokens WHERE token = $1 AND user_id = $2`, token, userID); err != nil {
		logger.Errorf(ctx, "failed to delete device token for user: %s", err)
		return domains.ErrDatabaseQuery
	}
	return nil
}
