package transaction

import (
	"sen1or/letslive/finance/domains"
)

type TransactionService struct {
	accountRepo     domains.AccountRepository
	transactionRepo domains.TransactionRepository
	currencyRepo    domains.CurrencyRepository
}

func NewTransactionService(accountRepo domains.AccountRepository, transactionRepo domains.TransactionRepository, currencyRepo domains.CurrencyRepository) *TransactionService {
	return &TransactionService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		currencyRepo:    currencyRepo,
	}
}
