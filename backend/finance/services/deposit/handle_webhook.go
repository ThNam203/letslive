package deposit

import (
	"context"
	"errors"
	"sen1or/letslive/finance/domains"
	gatewaypayment "sen1or/letslive/finance/gateway/payment"
	"sen1or/letslive/shared/pkg/logger"
)

// HandleWebhook verifies the provider signature, then completes or fails the
// payment idempotently. On completion the user wallet is credited and the
// escrow account is debited inside a single DB transaction; the zero-sum
// trigger validates the ledger on the status transition.
func (s *DepositService) HandleWebhook(ctx context.Context, providerName domains.PaymentProvider, payload []byte, signature string) error {
	gateway, ok := s.gateways[providerName]
	if !ok {
		return domains.ErrInvalidInput
	}

	event, err := gateway.VerifyWebhook(payload, signature)
	if err != nil {
		logger.Errorf(ctx, "webhook verify failed [handlewebhook: %v]", err)
		return domains.ErrUnauthorized
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
		if errResp := s.paymentRepo.UpdateStatus(ctx, payment.Id, domains.ProcessStatusFailed); errResp != nil {
			return errResp
		}
		s.failTransaction(ctx, payment.TransactionId)
		return nil

	case gatewaypayment.WebhookEventCompleted:
		tx, txErr := s.transactionRepo.GetById(ctx, payment.TransactionId)
		if txErr != nil {
			return txErr
		}

		// crash recovery: ledger already completed but the payment row was not
		// marked before a previous attempt died — just finish the payment
		if tx.Status == domains.ProcessStatusCompleted {
			return s.paymentRepo.UpdateStatus(ctx, payment.Id, domains.ProcessStatusCompleted)
		}

		if tx.ActorId == nil {
			logger.Errorf(ctx, "completed webhook for transaction %s missing actor_id", tx.Id)
			return domains.ErrTransactionFailed
		}

		userAccount, errResp := s.accountRepo.GetUserWalletByOwnerId(ctx, *tx.ActorId)
		if errResp != nil {
			if !errors.Is(errResp, domains.ErrAccountNotFound) {
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
