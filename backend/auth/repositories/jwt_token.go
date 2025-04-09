package repositories

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	serviceresponse "sen1or/letslive/auth/responses"

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

func (r *postgresRefreshTokenRepo) Update(ctx context.Context, token *domains.RefreshToken) *serviceresponse.ServiceErrorResponse {
	result, err := r.dbConn.Exec(ctx, `
		UPDATE refresh_tokens 
		SET revoked_at = $1 
		WHERE token = $2
	`, &token.ExpiresAt, &token.Token)
	if err != nil {
		return serviceresponse.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		return serviceresponse.ErrRefreshTokenNotFound
	}

	return nil
}

func (r *postgresRefreshTokenRepo) RevokeAllTokensOfUser(ctx context.Context, userID uuid.UUID) *serviceresponse.ServiceErrorResponse {
	var timeNow = time.Now()
	_, err := r.dbConn.Exec(ctx, `
		UPDATE refresh_tokens 
		SET revoked_at = $1 
		WHERE user_id = $2
	`, &timeNow, userID.String())
	if err != nil {
		return serviceresponse.ErrDatabaseQuery
	}

	return nil
}

func (r *postgresRefreshTokenRepo) Insert(ctx context.Context, tokenRecord *domains.RefreshToken) *serviceresponse.ServiceErrorResponse {
	params := pgx.NamedArgs{
		"token":      tokenRecord.Token,
		"expires_at": tokenRecord.ExpiresAt,
		"user_id":    tokenRecord.UserId,
	}

	result, err := r.dbConn.Exec(ctx, `
		INSERT INTO refresh_tokens (
			token, 
			expires_at, 
			user_id
		) values (
			@token, 
			@expires_at, 
			@user_id
		)
	`, params)

	if err != nil {
		return serviceresponse.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		return serviceresponse.ErrInternalServer
	}

	return nil
}

func (r *postgresRefreshTokenRepo) FindByValue(ctx context.Context, tokenVal string) (*domains.RefreshToken, *serviceresponse.ServiceErrorResponse) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT * 
		FROM refresh_tokens 
		WHERE token = $1
	`, tokenVal)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, serviceresponse.ErrRefreshTokenNotFound
		}
		return nil, serviceresponse.ErrDatabaseQuery
	}
	defer rows.Close()

	token, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.RefreshToken])
	if err != nil {
		return nil, serviceresponse.ErrDatabaseIssue
	}

	return &token, nil
}
