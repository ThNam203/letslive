package repositories

import (
	"context"
	"errors"
	"sen1or/lets-live/auth/domains"
	servererrors "sen1or/lets-live/auth/errors"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VerifyTokenRepository interface {
	Insert(newToken domains.VerifyToken) *servererrors.ServerError
	GetByValue(token string) (*domains.VerifyToken, *servererrors.ServerError)
	DeleteByID(uuid.UUID) *servererrors.ServerError
	DeleteByValue(token string) *servererrors.ServerError
	DeleteAllOfUser(userId uuid.UUID) *servererrors.ServerError
}

type postgresVerifyTokenRepo struct {
	dbConn *pgxpool.Pool
}

func NewVerifyTokenRepo(conn *pgxpool.Pool) VerifyTokenRepository {
	return &postgresVerifyTokenRepo{
		dbConn: conn,
	}
}

func (r *postgresVerifyTokenRepo) Insert(newToken domains.VerifyToken) *servererrors.ServerError {
	_ = pgx.NamedArgs{
		"token":      newToken.Token,
		"expires_at": newToken.ExpiresAt,
		"user_id":    newToken.UserID,
	}

	result, err := r.dbConn.Exec(context.Background(), "insert into verify_tokens (token, expires_at, user_id) values ($1, $2, $3)", newToken.Token, newToken.ExpiresAt, newToken.UserID)
	if err != nil {
		return servererrors.ErrDatabaseQuery
	} else if result.RowsAffected() == 0 {
		return servererrors.ErrDatabaseIssue
	}

	return nil
}

func (r *postgresVerifyTokenRepo) GetByValue(token string) (*domains.VerifyToken, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), "select * from verify_tokens where token = $1", token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrVerifyTokenNotFound
		}

		return nil, servererrors.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.VerifyToken])
	if err != nil {
		return nil, servererrors.ErrDatabaseIssue
	}

	return &user, nil
}

func (r *postgresVerifyTokenRepo) DeleteByID(tokenID uuid.UUID) *servererrors.ServerError {
	result, err := r.dbConn.Exec(context.Background(), "DELETE FROM verify_tokens WHERE id = $1", tokenID.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return servererrors.ErrVerifyTokenNotFound
		}

		return servererrors.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		return servererrors.ErrVerifyTokenNotFound
	}

	return nil
}

func (r *postgresVerifyTokenRepo) DeleteByValue(token string) *servererrors.ServerError {
	result, err := r.dbConn.Exec(context.Background(), "DELETE FROM verify_tokens WHERE token = $1", token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return servererrors.ErrVerifyTokenNotFound
		}

		return servererrors.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		return servererrors.ErrVerifyTokenNotFound
	}

	return nil
}

func (r *postgresVerifyTokenRepo) DeleteAllOfUser(userId uuid.UUID) *servererrors.ServerError {
	result, err := r.dbConn.Exec(context.Background(), "DELETE FROM verify_tokens WHERE user_id = $1", userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return servererrors.ErrAuthNotFound
		}

		return servererrors.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		return servererrors.ErrAuthNotFound
	}

	return nil
}
