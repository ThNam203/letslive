package currency

import (
	"sen1or/letslive/finance/domains"
)

type CurrencyService struct {
	currencyRepo domains.CurrencyRepository
}

func NewCurrencyService(currencyRepo domains.CurrencyRepository) *CurrencyService {
	return &CurrencyService{
		currencyRepo: currencyRepo,
	}
}
