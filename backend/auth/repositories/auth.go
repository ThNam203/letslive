package repositories

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/pkg/logger"
	serviceresponse "sen1or/letslive/auth/responses"

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

func (r *postgresAuthRepo) GetByID(ctx context.Context, authID uuid.UUID) (*domains.Auth, *serviceresponse.ServiceErrorResponse) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT * 
		FROM auths 
		WHERE id = $1
	`, authID.String())
	if err != nil {
		logger.Errorf("failed to get auth from id: %s", err)
		return nil, serviceresponse.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, serviceresponse.ErrAuthNotFound
		}

		logger.Errorf("failed to collect row: %s", err)
		return nil, serviceresponse.ErrDatabaseIssue
	}

	return &user, nil
}

func (r *postgresAuthRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domains.Auth, *serviceresponse.ServiceErrorResponse) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT * 
		FROM auths 
		WHERE user_id = $1
	`, userID.String())
	if err != nil {
		logger.Errorf("failed to get auth from user id: %s", err)
		return nil, serviceresponse.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, serviceresponse.ErrAuthNotFound
		}

		logger.Errorf("failed to collect row: %s", err)
		return nil, serviceresponse.ErrDatabaseIssue
	}

	return &user, nil
}

func (r *postgresAuthRepo) GetByEmail(ctx context.Context, email string) (*domains.Auth, *serviceresponse.ServiceErrorResponse) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT * 
		FROM auths 
		WHERE email = $1
	`, email)
	if err != nil {
		logger.Errorf("failed to get auth from email: %s", err)
		return nil, serviceresponse.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, serviceresponse.ErrAuthNotFound
		}

		logger.Errorf("failed to collect row: %s", err)
		return nil, serviceresponse.ErrDatabaseIssue
	}

	return &user, nil
}

func (r *postgresAuthRepo) Create(ctx context.Context, newAuth domains.Auth) (*domains.Auth, *serviceresponse.ServiceErrorResponse) {
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
		logger.Errorf("failed to create auth: %s", err)
		return nil, serviceresponse.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, serviceresponse.ErrAuthNotFound
		}

		logger.Errorf("failed to collect row: %s", err)
		return nil, serviceresponse.ErrDatabaseIssue
	}

	return &user, nil
}

func (r *postgresAuthRepo) UpdatePasswordHash(ctx context.Context, authId, newPasswordHash string) *serviceresponse.ServiceErrorResponse {
	result, err := r.dbConn.Exec(ctx, `
		UPDATE auths 
		SET password_hash = $1 
		WHERE id = $2
	`, newPasswordHash, authId)
	if err != nil {
		logger.Errorf("failed to update password hash: %s", err)
		return serviceresponse.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		return serviceresponse.ErrAuthNotFound
	}

	return nil
}
