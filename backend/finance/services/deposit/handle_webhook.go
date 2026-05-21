package deposit

import (
	"context"
	"sen1or/letslive/finance/domains"
	gatewaypayment "sen1or/letslive/finance/gateway/payment"
	response "sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/logger"
)

// HandleWebhook verifies the provider signature, then completes or fails the
// payment idempotently. On completion the user wallet is credited and the
// escrow account is debited inside a single DB transaction; the zero-sum
// trigger validates the ledger on the status transition.
func (s *DepositService) HandleWebhook(ctx context.Context, providerName domains.PaymentProvider, payload []byte, signature string) *response.Response[any] {
	gateway, ok := s.gateways[providerName]
	if !ok {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		)
	}

	event, err := gateway.VerifyWebhook(payload, signature)
	if err != nil {
		logger.Errorf(ctx, "webhook verify failed [handlewebhook: %v]", err)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		)
	}
	if event.Type == gatewaypayment.WebhookEventIgnored {
		return nil
	}

	payment, errResp := s.paymentRepo.GetByProviderRef(ctx, providerName, event.ProviderRef)
	if errResp != nil {
		return errResp
	}

	// idempotent: terminal states stay terminal
	if payment.Status == domains.ProcessStatusCompleted || payment.Status == domains.ProcessStatusFailed {
		return nil
	}

	switch event.Type {
	case gatewaypayment.WebhookEventFailed:
		return s.paymentRepo.UpdateStatus(ctx, payment.Id, domains.ProcessStatusFailed)

	case gatewaypayment.WebhookEventCompleted:
		tx, txErr := s.transactionRepo.GetById(ctx, payment.TransactionId)
		if txErr != nil {
			return txErr
		}
		if tx.ActorId == nil {
			logger.Errorf(ctx, "completed webhook for transaction %s missing actor_id", tx.Id)
			return response.NewResponseFromTemplate[any](
				response.RES_ERR_TRANSACTION_FAILED,
				nil,
				nil,
				nil,
			)
		}

		userAccount, errResp := s.accountRepo.GetUserWalletByOwnerId(ctx, *tx.ActorId)
		if errResp != nil {
			if errResp.Code != response.RES_ERR_ACCOUNT_NOT_FOUND_CODE {
				return errResp
			}
			created, createErr := s.accountRepo.CreateUserWallet(ctx, *tx.ActorId)
			if createErr != nil {
				return createErr
			}
			userAccount = created
		}

		escrow, errResp := s.accountRepo.GetEscrow(ctx)
		if errResp != nil {
			return errResp
		}

		entries := []domains.LedgerEntryDraft{
			{AccountId: userAccount.Id, CurrencyCode: payment.CurrencyCode, Amount: payment.Amount},
			{AccountId: escrow.Id, CurrencyCode: payment.CurrencyCode, Amount: -payment.Amount},
		}

		if completeErr := s.transactionRepo.CompleteWithEntries(ctx, tx.Id, entries); completeErr != nil {
			return completeErr
		}

		return s.paymentRepo.UpdateStatus(ctx, payment.Id, domains.ProcessStatusCompleted)
	}

	return nil
}
