package deposit

import (
	"sen1or/letslive/finance/domains"
	gatewaypayment "sen1or/letslive/finance/gateway/payment"
)

type DepositService struct {
	accountRepo     domains.AccountRepository
	currencyRepo    domains.CurrencyRepository
	transactionRepo domains.TransactionRepository
	paymentRepo     domains.PaymentRepository
	gateways        map[domains.PaymentProvider]gatewaypayment.PaymentGateway
	minAmount       int64
	maxAmount       int64
}

func NewDepositService(
	accountRepo domains.AccountRepository,
	currencyRepo domains.CurrencyRepository,
	transactionRepo domains.TransactionRepository,
	paymentRepo domains.PaymentRepository,
	gateways []gatewaypayment.PaymentGateway,
	minAmount int64,
	maxAmount int64,
) *DepositService {
	indexed := make(map[domains.PaymentProvider]gatewaypayment.PaymentGateway, len(gateways))
	for _, g := range gateways {
		indexed[g.Provider()] = g
	}
	return &DepositService{
		accountRepo:     accountRepo,
		currencyRepo:    currencyRepo,
		transactionRepo: transactionRepo,
		paymentRepo:     paymentRepo,
		gateways:        indexed,
		minAmount:       minAmount,
		maxAmount:       maxAmount,
	}
}
