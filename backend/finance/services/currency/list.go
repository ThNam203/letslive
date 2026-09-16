package currency

import (
	"context"
	"sen1or/letslive/finance/domains"
)

func (s *CurrencyService) List(ctx context.Context) ([]domains.Currency, error) {
	return s.currencyRepo.List(ctx)
}
