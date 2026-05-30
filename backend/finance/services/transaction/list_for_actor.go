package transaction

import (
	"context"
	"sen1or/letslive/finance/dto"
	response "sen1or/letslive/finance/response"

	"github.com/gofrs/uuid/v5"
)

func (s *TransactionService) ListForActor(ctx context.Context, actorId uuid.UUID, page int, limit int) ([]dto.TransactionResponse, int, *response.Response[any]) {
	account, errResp := s.accountRepo.GetUserWalletByOwnerId(ctx, actorId)
	if errResp != nil && errResp.Code != response.RES_ERR_ACCOUNT_NOT_FOUND_CODE {
		return nil, 0, errResp
	}

	txs, total, errResp := s.transactionRepo.ListByActor(ctx, actorId, page, limit)
	if errResp != nil {
		return nil, 0, errResp
	}

	currencies, errResp := s.currencyRepo.List(ctx)
	if errResp != nil {
		return nil, 0, errResp
	}
	precisionByCode := make(map[string]int, len(currencies))
	for _, c := range currencies {
		precisionByCode[c.Code] = c.Precision
	}

	out := make([]dto.TransactionResponse, 0, len(txs))
	for _, tx := range txs {
		entries := []dto.LedgerEntryResponse{}
		if account != nil {
			rawEntries, entryErr := s.transactionRepo.GetEntriesForAccount(ctx, tx.Id, account.Id)
			if entryErr != nil {
				return nil, 0, entryErr
			}
			for _, e := range rawEntries {
				entries = append(entries, dto.NewLedgerEntryResponse(e, precisionByCode[e.CurrencyCode]))
			}
		}
		out = append(out, dto.TransactionResponse{
			Transaction: tx,
			Entries:     entries,
		})
	}
	return out, total, nil
}
