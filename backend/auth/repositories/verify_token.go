package repositories

import (
	"context"
	"sen1or/lets-live/auth/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VerifyTokenRepository interface {
	Create(newToken domains.VerifyToken) error
	GetByValue(token string) (*domains.VerifyToken, error)
	DeleteByID(uuid.UUID) error
	DeleteByValue(token string) error
	DeleteAllOfUser(userId uuid.UUID) error
}

type postgresVerifyTokenRepo struct {
	dbConn *pgxpool.Pool
}

func NewVerifyTokenRepo(conn *pgxpool.Pool) VerifyTokenRepository {
	return &postgresVerifyTokenRepo{
		dbConn: conn,
	}
}

func (r *postgresVerifyTokenRepo) Create(newToken domains.VerifyToken) error {
	_ = pgx.NamedArgs{
		"token":      newToken.Token,
		"expires_at": newToken.ExpiresAt,
		"user_id":    newToken.UserID,
	}

	_, err := r.dbConn.Exec(context.Background(), "insert into verify_tokens (token, expires_at, user_id) values ($1, $2, $3)", newToken.Token, newToken.ExpiresAt, newToken.UserID)

	return err
}

func (r *postgresVerifyTokenRepo) GetByValue(token string) (*domains.VerifyToken, error) {
	rows, err := r.dbConn.Query(context.Background(), "select * from verify_tokens where token = $1", token)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.VerifyToken])

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *postgresVerifyTokenRepo) DeleteByID(tokenID uuid.UUID) error {
	_, err := r.dbConn.Exec(context.Background(), "DELETE FROM verify_tokens WHERE id = $1", tokenID.String())
	return err
}

func (r *postgresVerifyTokenRepo) DeleteByValue(token string) error {
	_, err := r.dbConn.Exec(context.Background(), "DELETE FROM verify_tokens WHERE token = $1", token)
	return err
}

func (r *postgresVerifyTokenRepo) DeleteAllOfUser(userId uuid.UUID) error {
	_, err := r.dbConn.Exec(context.Background(), "DELETE FROM verify_tokens WHERE user_id = $1", userId)
	return err
}
