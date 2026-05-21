package payment

import (
	"sen1or/letslive/finance/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresPaymentRepo struct {
	dbConn *pgxpool.Pool
}

func NewPaymentRepository(conn *pgxpool.Pool) domains.PaymentRepository {
	return &postgresPaymentRepo{
		dbConn: conn,
	}
}
