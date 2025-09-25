package repositories

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/pkg/logger"
	serviceresponse "sen1or/letslive/auth/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresAuthRepo struct {
	dbConn *pgxpool.Pool
}

func NewAuthRepository(conn *pgxpool.Pool) domains.AuthRepository {
	return &postgresAuthRepo{
		dbConn: conn,
	}
}

func (r *postgresAuthRepo) GetByID(ctx context.Context, authId uuid.UUID) (*domains.Auth, *serviceresponse.Response[any]) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT * 
		FROM auths 
		WHERE id = $1
	`, authId.String())
	if err != nil {
		logger.Errorf(ctx, "failed to get auth from id: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"authId": authId}},
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
				&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"authId": authId}},
			)
		}

		logger.Errorf(ctx, "failed to collect row: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"authId": authId}},
		)
	}

	return &user, nil
}

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

func (r *postgresAuthRepo) UpdatePasswordHash(ctx context.Context, authId, newPasswordHash string) *serviceresponse.Response[any] {
	result, err := r.dbConn.Exec(ctx, `
		UPDATE auths 
		SET password_hash = $1 
		WHERE id = $2
	`, newPasswordHash, authId)
	if err != nil {
		logger.Errorf(ctx, "failed to update password hash: %s", err)
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"authId": authId}},
		)
	}

	if result.RowsAffected() == 0 {
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_AUTH_NOT_FOUND,
			nil,
			nil,
			&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"authId": authId}},
		)
	}

	return nil
}
