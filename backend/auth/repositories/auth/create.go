package auth

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/pkg/logger"
	serviceresponse "sen1or/letslive/auth/response"

	"github.com/jackc/pgx/v5"
)

func (r *postgresAuthRepo) Create(ctx context.Context, newAuth domains.Auth) (*domains.Auth, *serviceresponse.Response[any]) {
	params := pgx.NamedArgs{
		"email":         newAuth.Email,
		"password_hash": newAuth.PasswordHash,
		"user_id":       newAuth.UserId,
	}

	rows, err := r.dbConn.Query(ctx, `
		INSERT INTO auths (
			email,
			password_hash,
			user_id
		) values (
			@email, 
			@password_hash, 
			@user_id
		) RETURNING *
	`, params)
	if err != nil {
		logger.Errorf(ctx, "failed to create auth: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"newAuth.email": newAuth.Email, "newAuth.userId": newAuth.UserId}},
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
				&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"newAuth.email": newAuth.Email, "newAuth.userId": newAuth.UserId}},
			)
		}

		logger.Errorf(ctx, "failed to collect row: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"newAuth.email": newAuth.Email, "newAuth.userId": newAuth.UserId}},
		)
	}

	return &user, nil
}
