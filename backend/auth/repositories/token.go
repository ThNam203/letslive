package repositories

import (
	"context"
	"sen1or/lets-live/auth/domains"

	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

type RefreshTokenRepository interface {
	RevokeAllTokensOfUser(userId uuid.UUID) error

	Create(*domains.RefreshToken) error
	FindByValue(string) (*domains.RefreshToken, error)
	Update(*domains.RefreshToken) error
}

type postgresRefreshTokenRepo struct {
	dbConn *pgx.Conn
}

func NewRefreshTokenRepository(conn *pgx.Conn) RefreshTokenRepository {
	return &postgresRefreshTokenRepo{
		dbConn: conn,
	}
}

func (r *postgresRefreshTokenRepo) Update(token *domains.RefreshToken) error {
	_, err := r.dbConn.Exec(context.Background(), "UPDATE refresh_tokens SET revoked_at = $1 WHERE id = $2", &token.ExpiresAt, &token.ID)
	return err
}

func (r *postgresRefreshTokenRepo) RevokeAllTokensOfUser(userID uuid.UUID) error {
	var timeNow = time.Now()
	_, err := r.dbConn.Exec(context.Background(), "UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2", &timeNow, userID.String())
	return err
}
func (r *postgresRefreshTokenRepo) Create(tokenRecord *domains.RefreshToken) error {
	params := pgx.NamedArgs{
		"value":      tokenRecord.Value,
		"expires_at": tokenRecord.ExpiresAt,
		"user_id":    tokenRecord.UserID,
	}

	_, err := r.dbConn.Exec(context.Background(), "insert into refresh_tokens (value, expires_at, user_id) values (@value, @expires_at, @user_id)", params)

	return err
}

func (r *postgresRefreshTokenRepo) FindByValue(tokenValue string) (*domains.RefreshToken, error) {
	rows, err := r.dbConn.Query(context.Background(), "select * from refresh_tokens where id = $1", tokenValue)
	if err != nil {
		return nil, err
	}

	token, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.RefreshToken])

	if err != nil {
		return nil, err
	}

	return &token, nil
}
