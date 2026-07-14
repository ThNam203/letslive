package deposit

import (
	"context"
	"sen1or/letslive/finance/domains"
	gatewaypayment "sen1or/letslive/finance/gateway/payment"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
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

// failTransaction best-effort marks a transaction failed so it does not dangle in
// 'created' forever; the caller is already on an error path, so only log on failure.
func (s *DepositService) failTransaction(ctx context.Context, id uuid.UUID) {
	if errResp := s.transactionRepo.UpdateStatus(ctx, id, domains.ProcessStatusFailed); errResp != nil {
		logger.Errorf(ctx, "failed to mark transaction %s as failed [failtransaction]", id)
	}
}
