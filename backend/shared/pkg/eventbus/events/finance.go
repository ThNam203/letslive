package events

import (
	"github.com/gofrs/uuid/v5"
)

// Finance event types.
const (
	PaymentCompleted = "finance.payment_completed"
	PaymentFailed    = "finance.payment_failed"
	DonationSent     = "finance.donation_sent"
)

// PaymentCompletedEvent is emitted when a payment is successfully processed.
type PaymentCompletedEvent struct {
	PaymentId     uuid.UUID `json:"paymentId"`
	TransactionId uuid.UUID `json:"transactionId"`
	UserId        uuid.UUID `json:"userId"`
	Amount        string    `json:"amount"`
	CurrencyCode  string    `json:"currencyCode"`
	Provider      string    `json:"provider"`
}

// PaymentFailedEvent is emitted when a payment fails.
type PaymentFailedEvent struct {
	PaymentId uuid.UUID `json:"paymentId"`
	UserId    uuid.UUID `json:"userId"`
	ErrorMsg  string    `json:"errorMsg"`
}

// DonationSentEvent is emitted when a donation transaction completes.
type DonationSentEvent struct {
	TransactionId uuid.UUID `json:"transactionId"`
	SenderId      uuid.UUID `json:"senderId"`
	ReceiverId    uuid.UUID `json:"receiverId"`
	Amount        string    `json:"amount"`
	CurrencyCode  string    `json:"currencyCode"`
}
