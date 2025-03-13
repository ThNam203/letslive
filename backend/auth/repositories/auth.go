package repositories

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	servererrors "sen1or/letslive/auth/errors"
	"sen1or/letslive/auth/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository interface {
	GetByID(uuid.UUID) (*domains.Auth, *servererrors.ServerError)
	GetByUserID(uuid.UUID) (*domains.Auth, *servererrors.ServerError)
	GetByEmail(string) (*domains.Auth, *servererrors.ServerError)

	Create(domains.Auth) (*domains.Auth, *servererrors.ServerError)
	UpdatePasswordHash(authId, newPasswordHash string) *servererrors.ServerError
	Delete(uuid.UUID) *servererrors.ServerError
}

type postgresAuthRepo struct {
	dbConn *pgxpool.Pool
}

func NewAuthRepository(conn *pgxpool.Pool) AuthRepository {
	return &postgresAuthRepo{
		dbConn: conn,
	}
}

func (r *postgresAuthRepo) GetByID(authID uuid.UUID) (*domains.Auth, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), "select * from auths where id = $1", authID.String())
	if err != nil {
		logger.Errorf("failed to get auth from id: %s", err)
		return nil, servererrors.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrAuthNotFound
		}

		logger.Errorf("failed to collect row: %s", err)
		return nil, servererrors.ErrDatabaseIssue
	}

	return &user, nil
}

func (r *postgresAuthRepo) GetByUserID(userID uuid.UUID) (*domains.Auth, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), "select * from auths where user_id = $1", userID.String())
	if err != nil {
		logger.Errorf("failed to get auth from user id: %s", err)
		return nil, servererrors.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrAuthNotFound
		}

		logger.Errorf("failed to collect row: %s", err)
		return nil, servererrors.ErrDatabaseIssue
	}

	return &user, nil
}

func (r *postgresAuthRepo) GetByEmail(email string) (*domains.Auth, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), "select * from auths where email = $1", email)
	if err != nil {
		logger.Errorf("failed to get auth from email: %s", err)
		return nil, servererrors.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrAuthNotFound
		}

		logger.Errorf("failed to collect row: %s", err)
		return nil, servererrors.ErrDatabaseIssue
	}

	return &user, nil
}

func (r *postgresAuthRepo) Create(newAuth domains.Auth) (*domains.Auth, *servererrors.ServerError) {
	params := pgx.NamedArgs{
		"email":         newAuth.Email,
		"password_hash": newAuth.PasswordHash,
		"user_id":       newAuth.UserId,
	}

	rows, err := r.dbConn.Query(context.Background(), "insert into auths (email, password_hash, user_id) values (@email, @password_hash, @user_id) RETURNING *", params)
	if err != nil {
		logger.Errorf("failed to create auth: %s", err)
		return nil, servererrors.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrAuthNotFound
		}

		logger.Errorf("failed to collect row: %s", err)
		return nil, servererrors.ErrDatabaseIssue
	}

	return &user, nil
}

func (r *postgresAuthRepo) UpdatePasswordHash(authId, newPasswordHash string) *servererrors.ServerError {
	result, err := r.dbConn.Exec(context.Background(), "UPDATE auths SET password_hash = $1 WHERE id = $2 RETURNING *", newPasswordHash, authId)
	if err != nil {
		logger.Errorf("failed to update password hash: %s", err)
		return servererrors.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		return servererrors.ErrAuthNotFound
	}

	return nil
}

func (r *postgresAuthRepo) Delete(userID uuid.UUID) *servererrors.ServerError {
	result, err := r.dbConn.Exec(context.Background(), "DELETE FROM auths WHERE id = $1", userID.String())
	if err != nil {
		logger.Errorf("failed to delete auth: %s", err)
		return servererrors.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		return servererrors.ErrAuthNotFound
	}

	return nil
}
