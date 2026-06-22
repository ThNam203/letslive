package currency

import (
	"context"
	"sen1or/letslive/finance/domains"
	response "sen1or/letslive/finance/response"
)

func (s *CurrencyService) List(ctx context.Context) ([]domains.Currency, *response.Response[any]) {
	return s.currencyRepo.List(ctx)
}
