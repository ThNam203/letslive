package domains

import (
	"context"
	response "sen1or/letslive/finance/response"
	"time"

	"github.com/gofrs/uuid/v5"
)

type PaymentProvider string

const (
	PaymentProviderStripe PaymentProvider = "stripe"
	PaymentProviderPayPal PaymentProvider = "paypal"
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
	Create(ctx context.Context, payment Payment) (*Payment, *response.Response[any])
	GetById(ctx context.Context, id uuid.UUID) (*Payment, *response.Response[any])
	GetByProviderRef(ctx context.Context, provider PaymentProvider, providerRef string) (*Payment, *response.Response[any])
	UpdateStatus(ctx context.Context, id uuid.UUID, status ProcessStatus) *response.Response[any]
	ListByActor(ctx context.Context, actorId uuid.UUID, page int, limit int) ([]Payment, int, *response.Response[any])
}
