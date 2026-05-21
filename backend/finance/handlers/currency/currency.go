package currency

import (
	"sen1or/letslive/finance/handlers/basehandler"
	"sen1or/letslive/finance/services/currency"
)

type CurrencyHandler struct {
	basehandler.BaseHandler
	currencyService *currency.CurrencyService
}

func NewCurrencyHandler(currencyService *currency.CurrencyService) *CurrencyHandler {
	return &CurrencyHandler{
		currencyService: currencyService,
	}
}
