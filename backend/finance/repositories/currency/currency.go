package currency

import (
	"sen1or/letslive/finance/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresCurrencyRepo struct {
	dbConn *pgxpool.Pool
}

func NewCurrencyRepository(conn *pgxpool.Pool) domains.CurrencyRepository {
	return &postgresCurrencyRepo{
		dbConn: conn,
	}
}
