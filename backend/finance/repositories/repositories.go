package repositories

import (
	"sen1or/letslive/finance/domains"
	accountrepo "sen1or/letslive/finance/repositories/account"
	currencyrepo "sen1or/letslive/finance/repositories/currency"
	paymentrepo "sen1or/letslive/finance/repositories/payment"
	transactionrepo "sen1or/letslive/finance/repositories/transaction"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewAccountRepository(conn *pgxpool.Pool) domains.AccountRepository {
	return accountrepo.NewAccountRepository(conn)
}

func NewCurrencyRepository(conn *pgxpool.Pool) domains.CurrencyRepository {
	return currencyrepo.NewCurrencyRepository(conn)
}

func NewTransactionRepository(conn *pgxpool.Pool) domains.TransactionRepository {
	return transactionrepo.NewTransactionRepository(conn)
}

func NewPaymentRepository(conn *pgxpool.Pool) domains.PaymentRepository {
	return paymentrepo.NewPaymentRepository(conn)
}
