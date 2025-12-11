package auth

import (
	"context"
	"sen1or/letslive/auth/pkg/logger"
	serviceresponse "sen1or/letslive/auth/response"
)

func (r *postgresAuthRepo) UpdatePasswordHash(ctx context.Context, authId, newPasswordHash string) *serviceresponse.Response[any] {
	result, err := r.dbConn.Exec(ctx, `
		UPDATE auths 
		SET password_hash = $1 
		WHERE id = $2
	`, newPasswordHash, authId)
	if err != nil {
		logger.Errorf(ctx, "failed to update password hash: %s", err)
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"authId": authId}},
		)
	}

	if result.RowsAffected() == 0 {
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_AUTH_NOT_FOUND,
			nil,
			nil,
			&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"authId": authId}},
		)
	}

	return nil
}
