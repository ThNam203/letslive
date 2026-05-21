package transaction

import (
	"context"
	"sen1or/letslive/finance/dto"
	response "sen1or/letslive/finance/response"

	"github.com/gofrs/uuid/v5"
)

func (s *TransactionService) GetForActor(ctx context.Context, transactionId uuid.UUID, actorId uuid.UUID) (*dto.TransactionResponse, *response.Response[any]) {
	tx, errResp := s.transactionRepo.GetById(ctx, transactionId)
	if errResp != nil {
		return nil, errResp
	}
	if tx.ActorId == nil || *tx.ActorId != actorId {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_TRANSACTION_FAILED,
			nil,
			nil,
			nil,
		)
	}

	account, accErr := s.accountRepo.GetUserWalletByOwnerId(ctx, actorId)
	if accErr != nil && accErr.Code != response.RES_ERR_ACCOUNT_NOT_FOUND_CODE {
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
