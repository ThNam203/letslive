package auth

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/jackc/pgx/v5"
)

func (r *postgresAuthRepo) GetByEmail(ctx context.Context, email string) (*domains.Auth, error) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT * 
		FROM auths 
		WHERE email = $1
	`, email)
	if err != nil {
		logger.Errorf(ctx, "failed to get auth from email: %s", err)
		return nil, domains.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrAuthNotFound
		}

		logger.Errorf(ctx, "failed to collect row: %s", err)
		return nil, domains.ErrDatabaseIssue
	}

	return &user, nil
}
