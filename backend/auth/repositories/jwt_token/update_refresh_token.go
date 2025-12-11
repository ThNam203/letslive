package jwt_token

import (
	"context"
	"sen1or/letslive/auth/domains"
	serviceresponse "sen1or/letslive/auth/response"
)

func (r *postgresRefreshTokenRepo) Update(ctx context.Context, token *domains.RefreshToken) *serviceresponse.Response[any] {
	result, err := r.dbConn.Exec(ctx, `
		UPDATE refresh_tokens 
		SET revoked_at = $1 
		WHERE token = $2
	`, &token.ExpiresAt, &token.Token)
	if err != nil {
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	if result.RowsAffected() == 0 {
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_REFRESH_TOKEN_NOT_FOUND,
			nil,
			nil,
			nil,
		)
	}

	return nil
}
