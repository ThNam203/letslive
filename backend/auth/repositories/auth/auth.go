package auth

import (
	"sen1or/letslive/auth/domains"

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
