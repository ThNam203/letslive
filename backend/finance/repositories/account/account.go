package account

import (
	"sen1or/letslive/finance/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresAccountRepo struct {
	dbConn *pgxpool.Pool
}

func NewAccountRepository(conn *pgxpool.Pool) domains.AccountRepository {
	return &postgresAccountRepo{
		dbConn: conn,
	}
}
