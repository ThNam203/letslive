package transaction

import (
	"sen1or/letslive/finance/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresTransactionRepo struct {
	dbConn *pgxpool.Pool
}

func NewTransactionRepository(conn *pgxpool.Pool) domains.TransactionRepository {
	return &postgresTransactionRepo{
		dbConn: conn,
	}
}
