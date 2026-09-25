package auth

import (
	"context"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/shared/pkg/logger"
)

func (r *postgresAuthRepo) UpdatePasswordHash(ctx context.Context, authId, newPasswordHash string) error {
	result, err := r.dbConn.Exec(ctx, `
		UPDATE auths 
		SET password_hash = $1 
		WHERE id = $2
	`, newPasswordHash, authId)
	if err != nil {
		logger.Errorf(ctx, "failed to update password hash for auth %s: %s", authId, err)
		return domains.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		return domains.ErrAuthNotFound
	}

	return nil
}
