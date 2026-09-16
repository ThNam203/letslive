package deposit

import (
	"context"
	"errors"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/dto"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
)

// Initiate validates the deposit request, ensures the user has an active wallet,
// creates a created-state purchase transaction and a pending payment row, then
// asks the payment gateway to create a checkout session.
func (s *DepositService) Initiate(ctx context.Context, actorId uuid.UUID, req dto.DepositRequestDTO) (*dto.DepositResponse, error) {
	gateway, ok := s.gateways[domains.PaymentProvider(req.Provider)]
	if !ok {
		return nil, domains.ErrInvalidInput
	}

	currency, errResp := s.currencyRepo.GetByCode(ctx, req.CurrencyCode)
	if errResp != nil {
		return nil, errResp
	}

	amount, err := dto.ParseAmount(req.Amount, currency.Precision)
	if err != nil {
		return nil, domains.ErrInvalidAmount
	}
	if amount < s.minAmount {
		return nil, domains.ErrInvalidAmount
	}
	if amount > s.maxAmount {
		return nil, domains.ErrDepositLimitExceeded
	}

	account, errResp := s.accountRepo.GetUserWalletByOwnerId(ctx, actorId)
	if errResp != nil {
		if !errors.Is(errResp, domains.ErrAccountNotFound) {
			return nil, errResp
		}
		created, createErr := s.accountRepo.CreateUserWallet(ctx, actorId)
		if createErr != nil {
			return nil, createErr
		}
		account = created
	}
	if account.Status == domains.AccountStatusFrozen {
		return nil, domains.ErrAccountFrozen
	}

	idempotencyKey, err := uuid.NewV4()
	if err != nil {
		logger.Errorf(ctx, "uuid generation failed [initiatedeposit: %v]", err)
		return nil, domains.ErrInternal
	}
	reference := idempotencyKey.String()
	tx, errResp := s.transactionRepo.Create(ctx, domains.Transaction{
		Type:      domains.TransactionTypeDeposit,
		Reference: &reference,
		Status:    domains.ProcessStatusCreated,
		ActorId:   &actorId,
	})
	if errResp != nil {
		return nil, errResp
	}

	session, gwErr := gateway.CreateCheckoutSession(ctx, reference, amount, currency.Code, map[string]string{
		"transactionId": tx.Id.String(),
		"userId":        actorId.String(),
	})
	if gwErr != nil {
		logger.Errorf(ctx, "gateway checkout session error [initiatedeposit: %v]", gwErr)
		s.failTransaction(ctx, tx.Id)
		return nil, domains.ErrPaymentFailed
	}

	payment, errResp := s.paymentRepo.Create(ctx, domains.Payment{
		Provider:      gateway.Provider(),
		ProviderRef:   session.ProviderRef,
		CurrencyCode:  currency.Code,
		Amount:        amount,
		Status:        domains.ProcessStatusCreated,
		TransactionId: tx.Id,
	})
	if errResp != nil {
		// the provider session exists but we have no payment row to complete it against;
		// keep the ref in the log for manual reconciliation
		logger.Errorf(ctx, "payment row create failed after checkout session %s created [initiatedeposit]", session.ProviderRef)
		s.failTransaction(ctx, tx.Id)
		return nil, errResp
	}

	return &dto.DepositResponse{
		Payment:     dto.NewPaymentResponse(*payment, currency.Precision),
		CheckoutURL: session.CheckoutURL,
	}, nil
}
