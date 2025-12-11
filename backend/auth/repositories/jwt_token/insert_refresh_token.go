package jwt_token

import (
	"context"
	"sen1or/letslive/auth/domains"
	serviceresponse "sen1or/letslive/auth/response"

	"github.com/jackc/pgx/v5"
)

func (r *postgresRefreshTokenRepo) Insert(ctx context.Context, tokenRecord *domains.RefreshToken) *serviceresponse.Response[any] {
	params := pgx.NamedArgs{
		"token":      tokenRecord.Token,
		"expires_at": tokenRecord.ExpiresAt,
		"user_id":    tokenRecord.UserId,
	}

	result, err := r.dbConn.Exec(ctx, `
		INSERT INTO refresh_tokens (
			token, 
			expires_at, 
			user_id
		) values (
			@token, 
			@expires_at, 
			@user_id
		)
	`, params)

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
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	return nil
}
