package auth

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresAuthRepo) GetByUserID(ctx context.Context, userId uuid.UUID) (*domains.Auth, error) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT * 
		FROM auths 
		WHERE user_id = $1
	`, userId.String())
	if err != nil {
		logger.Errorf(ctx, "failed to get auth from user id %s: %s", userId, err)
		return nil, domains.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrAuthNotFound
		}

		logger.Errorf(ctx, "failed to collect auth row for user id %s: %s", userId, err)
		return nil, domains.ErrDatabaseIssue
	}

	return &user, nil
}
