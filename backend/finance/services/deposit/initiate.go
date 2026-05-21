package deposit

import (
	"context"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/dto"
	response "sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
)

// Initiate validates the deposit request, ensures the user has an active wallet,
// creates a created-state purchase transaction and a pending payment row, then
// asks the payment gateway to create a checkout session.
func (s *DepositService) Initiate(ctx context.Context, actorId uuid.UUID, req dto.DepositRequestDTO) (*dto.DepositResponse, *response.Response[any]) {
	gateway, ok := s.gateways[domains.PaymentProvider(req.Provider)]
	if !ok {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		)
	}

	currency, errResp := s.currencyRepo.GetByCode(ctx, req.CurrencyCode)
	if errResp != nil {
		return nil, errResp
	}

	amount, err := dto.ParseAmount(req.Amount, currency.Precision)
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_AMOUNT,
			nil,
			nil,
			nil,
		)
	}
	if amount < s.minAmount {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_AMOUNT,
			nil,
			nil,
			nil,
		)
	}
	if amount > s.maxAmount {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DEPOSIT_LIMIT_EXCEEDED,
			nil,
			nil,
			nil,
		)
	}

	account, errResp := s.accountRepo.GetUserWalletByOwnerId(ctx, actorId)
	if errResp != nil {
		if errResp.Code != response.RES_ERR_ACCOUNT_NOT_FOUND_CODE {
			return nil, errResp
		}
		created, createErr := s.accountRepo.CreateUserWallet(ctx, actorId)
		if createErr != nil {
			return nil, createErr
		}
		account = created
	}
	if account.Status == domains.AccountStatusFrozen {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_ACCOUNT_FROZEN,
			nil,
			nil,
			nil,
		)
	}

	idempotencyKey, err := uuid.NewV4()
	if err != nil {
		logger.Errorf(ctx, "uuid generation failed [initiatedeposit: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}
	reference := idempotencyKey.String()
	tx, errResp := s.transactionRepo.Create(ctx, domains.Transaction{
		Type:      domains.TransactionTypePurchase,
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
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_PAYMENT_FAILED,
			nil,
			nil,
			nil,
		)
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
		return nil, errResp
	}

	return &dto.DepositResponse{
		Payment:     dto.NewPaymentResponse(*payment, currency.Precision),
		CheckoutURL: session.CheckoutURL,
	}, nil
}
