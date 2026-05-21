package wallet

import (
	"context"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/dto"
	response "sen1or/letslive/finance/response"

	"github.com/gofrs/uuid/v5"
)

// GetOrCreateWallet fetches the user's wallet account; if absent, creates an active one.
// It then returns balances for every supported currency, defaulting missing rows to "0".
func (s *WalletService) GetOrCreateWallet(ctx context.Context, ownerId uuid.UUID) (*dto.WalletResponse, *response.Response[any]) {
	account, errResp := s.accountRepo.GetUserWalletByOwnerId(ctx, ownerId)
	if errResp != nil {
		if errResp.Code != response.RES_ERR_ACCOUNT_NOT_FOUND_CODE {
			return nil, errResp
		}
		created, createErr := s.accountRepo.CreateUserWallet(ctx, ownerId)
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

	balances, balErr := s.accountRepo.GetBalances(ctx, account.Id)
	if balErr != nil {
		return nil, balErr
	}

	currencies, curErr := s.currencyRepo.List(ctx)
	if curErr != nil {
		return nil, curErr
	}

	balanceByCurrency := make(map[string]domains.AccountBalance, len(balances))
	for _, b := range balances {
		balanceByCurrency[b.CurrencyCode] = b
	}

	out := make([]dto.BalanceResponse, 0, len(currencies))
	for _, c := range currencies {
		if existing, ok := balanceByCurrency[c.Code]; ok {
			out = append(out, dto.NewBalanceResponse(existing, c.Precision))
			continue
		}
		out = append(out, dto.BalanceResponse{
			AccountId:    account.Id.String(),
			CurrencyCode: c.Code,
			Balance:      dto.FormatAmount(0, c.Precision),
			LastEntryId:  nil,
		})
	}

	return &dto.WalletResponse{
		Account:  *account,
		Balances: out,
	}, nil
}
