package jwt_token

import (
	"context"
	"sen1or/letslive/auth/domains"
)

func (r *postgresRefreshTokenRepo) Update(ctx context.Context, token *domains.RefreshToken) error {
	result, err := r.dbConn.Exec(ctx, `
		UPDATE refresh_tokens 
		SET revoked_at = $1 
		WHERE token = $2
	`, &token.ExpiresAt, &token.Token)
	if err != nil {
		return domains.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		return domains.ErrRefreshTokenNotFound
	}

	return nil
}
