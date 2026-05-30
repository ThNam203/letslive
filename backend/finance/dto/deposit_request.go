package dto

type DepositRequestDTO struct {
	Provider     string `json:"provider" validate:"required,oneof=stripe paypal mock"`
	CurrencyCode string `json:"currencyCode" validate:"required"`
	Amount       string `json:"amount" validate:"required"`
}

type DepositResponse struct {
	Payment     PaymentResponse `json:"payment"`
	CheckoutURL string          `json:"checkoutUrl"`
}
