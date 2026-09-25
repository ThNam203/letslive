package auth

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresAuthRepo) GetByID(ctx context.Context, authId uuid.UUID) (*domains.Auth, error) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT * 
		FROM auths 
		WHERE id = $1
	`, authId.String())
	if err != nil {
		logger.Errorf(ctx, "failed to get auth from id %s: %s", authId, err)
		return nil, domains.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrAuthNotFound
		}

		logger.Errorf(ctx, "failed to collect auth row for id %s: %s", authId, err)
		return nil, domains.ErrDatabaseIssue
	}

	return &user, nil
}
