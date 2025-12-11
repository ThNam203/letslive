package auth

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/pkg/logger"
	serviceresponse "sen1or/letslive/auth/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresAuthRepo) GetByUserID(ctx context.Context, userId uuid.UUID) (*domains.Auth, *serviceresponse.Response[any]) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT * 
		FROM auths 
		WHERE user_id = $1
	`, userId.String())
	if err != nil {
		logger.Errorf(ctx, "failed to get auth from user id: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"userId": userId}},
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
				&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"userId": userId}},
			)
		}

		logger.Errorf(ctx, "failed to collect row: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"userId": userId}},
		)
	}

	return &user, nil
}
