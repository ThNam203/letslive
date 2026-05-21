package dto

import (
	"sen1or/letslive/finance/domains"
)

type LedgerEntryResponse struct {
	Id            string `json:"id"`
	TransactionId string `json:"transactionId"`
	AccountId     string `json:"accountId"`
	CurrencyCode  string `json:"currencyCode"`
	Amount        string `json:"amount"`
	CreatedAt     string `json:"createdAt"`
}

type TransactionResponse struct {
	domains.Transaction
	Entries []LedgerEntryResponse `json:"entries"`
}

func NewLedgerEntryResponse(e domains.LedgerEntry, precision int) LedgerEntryResponse {
	return LedgerEntryResponse{
		Id:            e.Id.String(),
		TransactionId: e.TransactionId.String(),
		AccountId:     e.AccountId.String(),
		CurrencyCode:  e.CurrencyCode,
		Amount:        FormatAmount(e.Amount, precision),
		CreatedAt:     e.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}
