package dto

import (
	"sen1or/letslive/finance/domains"
)

type PaymentResponse struct {
	Id                string `json:"id"`
	TransactionId     string `json:"transactionId"`
	Provider          string `json:"provider"`
	ProviderReference string `json:"providerReference"`
	CurrencyCode      string `json:"currencyCode"`
	Amount            string `json:"amount"`
	Status            string `json:"status"`
	CreatedAt         string `json:"createdAt"`
}

func NewPaymentResponse(p domains.Payment, precision int) PaymentResponse {
	return PaymentResponse{
		Id:                p.Id.String(),
		TransactionId:     p.TransactionId.String(),
		Provider:          string(p.Provider),
		ProviderReference: p.ProviderRef,
		CurrencyCode:      p.CurrencyCode,
		Amount:            FormatAmount(p.Amount, precision),
		Status:            string(p.Status),
		CreatedAt:         p.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}
