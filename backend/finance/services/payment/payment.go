package payment

import (
	"sen1or/letslive/finance/domains"
)

type PaymentService struct {
	paymentRepo     domains.PaymentRepository
	transactionRepo domains.TransactionRepository
	currencyRepo    domains.CurrencyRepository
}

func NewPaymentService(paymentRepo domains.PaymentRepository, transactionRepo domains.TransactionRepository, currencyRepo domains.CurrencyRepository) *PaymentService {
	return &PaymentService{
		paymentRepo:     paymentRepo,
		transactionRepo: transactionRepo,
		currencyRepo:    currencyRepo,
	}
}
