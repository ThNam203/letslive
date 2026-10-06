package adminaccount

import (
	"sen1or/letslive/admin/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresAdminAccountRepo struct {
	dbConn *pgxpool.Pool
}

func NewAdminAccountRepository(conn *pgxpool.Pool) domains.AdminAccountRepository {
	return &postgresAdminAccountRepo{dbConn: conn}
}
