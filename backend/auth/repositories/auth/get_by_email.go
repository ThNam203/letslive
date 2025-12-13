package auth

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/pkg/logger"
	serviceresponse "sen1or/letslive/auth/response"

	"github.com/jackc/pgx/v5"
)

func (r *postgresAuthRepo) GetByEmail(ctx context.Context, email string) (*domains.Auth, *serviceresponse.Response[any]) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT * 
		FROM auths 
		WHERE email = $1
	`, email)
	if err != nil {
		logger.Errorf(ctx, "failed to get auth from email: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"email": email}},
		)
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, serviceresponse.NewResponseFromTemplate[any](
				serviceresponse.RES_ERR_AUTH_NOT_FOUND,
				nil,
				nil,
				&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"email": email}},
			)
		}

		logger.Errorf(ctx, "failed to collect row: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"email": email}},
		)
	}

	return &user, nil
}
