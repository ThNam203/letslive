package domains

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

type PaymentProvider string

const (
	PaymentProviderStripe PaymentProvider = "stripe"
	// mock is registered on the dev profile only; see cmd/main.go
	PaymentProviderMock PaymentProvider = "mock"
)

type Payment struct {
	Id            uuid.UUID       `json:"id" db:"id"`
	Provider      PaymentProvider `json:"provider" db:"provider"`
	ProviderRef   string          `json:"providerReference" db:"provider_ref"`
	CurrencyCode  string          `json:"currencyCode" db:"currency_code"`
	Amount        int64           `json:"-" db:"amount"`
	Status        ProcessStatus   `json:"status" db:"status"`
	TransactionId uuid.UUID       `json:"transactionId" db:"transaction_id"`
	CreatedAt     time.Time       `json:"createdAt" db:"created_at"`
}

type PaymentRepository interface {
	Create(ctx context.Context, payment Payment) (*Payment, error)
	GetById(ctx context.Context, id uuid.UUID) (*Payment, error)
	GetByProviderRef(ctx context.Context, provider PaymentProvider, providerRef string) (*Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status ProcessStatus) error
	ListByActor(ctx context.Context, actorId uuid.UUID, page int, limit int) ([]Payment, int, error)
}
