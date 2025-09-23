package repositories

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	serviceresponse "sen1or/letslive/auth/response"

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

func (r *postgresRefreshTokenRepo) Update(ctx context.Context, token *domains.RefreshToken) *serviceresponse.Response[any] {
	result, err := r.dbConn.Exec(ctx, `
		UPDATE refresh_tokens 
		SET revoked_at = $1 
		WHERE token = $2
	`, &token.ExpiresAt, &token.Token)
	if err != nil {
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	if result.RowsAffected() == 0 {
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_REFRESH_TOKEN_NOT_FOUND,
			nil,
			nil,
			nil,
		)
	}

	return nil
}

func (r *postgresRefreshTokenRepo) RevokeAllTokensOfUser(ctx context.Context, userId uuid.UUID) *serviceresponse.Response[any] {
	var timeNow = time.Now()
	_, err := r.dbConn.Exec(ctx, `
		UPDATE refresh_tokens 
		SET revoked_at = $1 
		WHERE user_id = $2
	`, &timeNow, userId.String())
	if err != nil {
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"userId": userId}},
		)
	}

	return nil
}

func (r *postgresRefreshTokenRepo) Insert(ctx context.Context, tokenRecord *domains.RefreshToken) *serviceresponse.Response[any] {
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
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	if result.RowsAffected() == 0 {
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	return nil
}

func (r *postgresRefreshTokenRepo) FindByValue(ctx context.Context, tokenVal string) (*domains.RefreshToken, *serviceresponse.Response[any]) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT * 
		FROM refresh_tokens 
		WHERE token = $1
	`, tokenVal)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, serviceresponse.NewResponseFromTemplate[any](
				serviceresponse.RES_ERR_REFRESH_TOKEN_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}
	defer rows.Close()

	token, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.RefreshToken])
	if err != nil {
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return &token, nil
}
