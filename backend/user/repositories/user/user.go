package user

import (
	"sen1or/letslive/user/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresUserRepo struct {
	dbConn *pgxpool.Pool
}

func NewUserRepository(conn *pgxpool.Pool) domains.UserRepository {
	return &postgresUserRepo{
		dbConn: conn,
	}
}
