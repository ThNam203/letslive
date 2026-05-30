package transaction

import (
	"sen1or/letslive/finance/handlers/basehandler"
	"sen1or/letslive/finance/services/transaction"
)

type TransactionHandler struct {
	basehandler.BaseHandler
	transactionService *transaction.TransactionService
}

func NewTransactionHandler(transactionService *transaction.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}
