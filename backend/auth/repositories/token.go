package repositories

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	servererrors "sen1or/letslive/auth/errors"

	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRefreshTokenRepo struct {
	dbConn *pgxpool.Pool
}

func NewRefreshTokenRepository(conn *pgxpool.Pool) domains.RefreshTokenRepository {
	return &postgresRefreshTokenRepo{
		dbConn: conn,
	}
}

func (r *postgresRefreshTokenRepo) Update(token *domains.RefreshToken) *servererrors.ServerError {
	result, err := r.dbConn.Exec(context.Background(), "UPDATE refresh_tokens SET revoked_at = $1 WHERE id = $2", &token.ExpiresAt, &token.ID)
	if err != nil {
		return servererrors.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		return servererrors.ErrRefreshTokenNotFound
	}

	return nil
}

func (r *postgresRefreshTokenRepo) RevokeAllTokensOfUser(userID uuid.UUID) *servererrors.ServerError {
	var timeNow = time.Now()
	_, err := r.dbConn.Exec(context.Background(), "UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2", &timeNow, userID.String())
	if err != nil {
		return servererrors.ErrDatabaseQuery
	}

	return nil
}

func (r *postgresRefreshTokenRepo) Insert(tokenRecord *domains.RefreshToken) *servererrors.ServerError {
	params := pgx.NamedArgs{
		"value":      tokenRecord.Value,
		"expires_at": tokenRecord.ExpiresAt,
		"user_id":    tokenRecord.UserID,
	}

	result, err := r.dbConn.Exec(context.Background(), "insert into refresh_tokens (value, expires_at, user_id) values (@value, @expires_at, @user_id)", params)

	if err != nil {
		return servererrors.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		return servererrors.ErrInternalServer
	}

	return nil
}

func (r *postgresRefreshTokenRepo) FindByValue(tokenValue string) (*domains.RefreshToken, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), "select * from refresh_tokens where id = $1", tokenValue)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrRefreshTokenNotFound
		}
		return nil, servererrors.ErrDatabaseQuery
	}
	defer rows.Close()

	token, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.RefreshToken])
	if err != nil {
		return nil, servererrors.ErrDatabaseIssue
	}

	return &token, nil
}
