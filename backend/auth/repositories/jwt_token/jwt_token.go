package jwt_token

import (
	"sen1or/letslive/auth/domains"

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
