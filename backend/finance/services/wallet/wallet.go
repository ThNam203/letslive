package wallet

import (
	"sen1or/letslive/finance/domains"
)

type WalletService struct {
	accountRepo  domains.AccountRepository
	currencyRepo domains.CurrencyRepository
}

func NewWalletService(accountRepo domains.AccountRepository, currencyRepo domains.CurrencyRepository) *WalletService {
	return &WalletService{
		accountRepo:  accountRepo,
		currencyRepo: currencyRepo,
	}
}
