package jwt_token

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"

	"github.com/jackc/pgx/v5"
)

func (r *postgresRefreshTokenRepo) FindByValue(ctx context.Context, tokenVal string) (*domains.RefreshToken, error) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT * 
		FROM refresh_tokens 
		WHERE token = $1
	`, tokenVal)
	if err != nil {
		return nil, domains.ErrDatabaseQuery
	}
	defer rows.Close()

	token, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.RefreshToken])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrRefreshTokenNotFound
		}
		return nil, domains.ErrDatabaseIssue
	}

	return &token, nil
}
