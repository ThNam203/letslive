package transaction

import (
	"context"
	"errors"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/dto"

	"github.com/gofrs/uuid/v5"
)

func (s *TransactionService) GetForActor(ctx context.Context, transactionId uuid.UUID, actorId uuid.UUID) (*dto.TransactionResponse, error) {
	tx, errResp := s.transactionRepo.GetById(ctx, transactionId)
	if errResp != nil {
		return nil, errResp
	}
	if tx.ActorId == nil || *tx.ActorId != actorId {
		return nil, domains.ErrTransactionFailed
	}

	account, accErr := s.accountRepo.GetUserWalletByOwnerId(ctx, actorId)
	if accErr != nil && !errors.Is(accErr, domains.ErrAccountNotFound) {
		return nil, accErr
	}

	currencies, curErr := s.currencyRepo.List(ctx)
	if curErr != nil {
		return nil, curErr
	}
	precisionByCode := make(map[string]int, len(currencies))
	for _, c := range currencies {
		precisionByCode[c.Code] = c.Precision
	}

	entries := []dto.LedgerEntryResponse{}
	if account != nil {
		rawEntries, entryErr := s.transactionRepo.GetEntriesForAccount(ctx, tx.Id, account.Id)
		if entryErr != nil {
			return nil, entryErr
		}
		for _, e := range rawEntries {
			entries = append(entries, dto.NewLedgerEntryResponse(e, precisionByCode[e.CurrencyCode]))
		}
	}

	return &dto.TransactionResponse{
		Transaction: *tx,
		Entries:     entries,
	}, nil
}
