package dto

import (
	"sen1or/letslive/finance/domains"
)

type BalanceResponse struct {
	AccountId    string  `json:"accountId"`
	CurrencyCode string  `json:"currencyCode"`
	Balance      string  `json:"balance"`
	LastEntryId  *string `json:"lastEntryId"`
}

type WalletResponse struct {
	Account  domains.Account   `json:"account"`
	Balances []BalanceResponse `json:"balances"`
}

func NewBalanceResponse(b domains.AccountBalance, precision int) BalanceResponse {
	var lastEntryId *string
	if b.LastEntryId != nil {
		s := b.LastEntryId.String()
		lastEntryId = &s
	}
	return BalanceResponse{
		AccountId:    b.AccountId.String(),
		CurrencyCode: b.CurrencyCode,
		Balance:      FormatAmount(b.Balance, precision),
		LastEntryId:  lastEntryId,
	}
}
