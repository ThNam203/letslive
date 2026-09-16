package jwt_token

import (
	"context"
	"sen1or/letslive/auth/domains"

	"github.com/jackc/pgx/v5"
)

func (r *postgresRefreshTokenRepo) Insert(ctx context.Context, tokenRecord *domains.RefreshToken) error {
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
		return domains.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		return domains.ErrInternal
	}

	return nil
}
