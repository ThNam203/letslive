package domains

import (
	"context"
)

type Currency struct {
	Code      string `json:"code" db:"code"`
	Name      string `json:"name" db:"name"`
	Precision int    `json:"precision" db:"precision"`
}

type CurrencyRepository interface {
	List(ctx context.Context) ([]Currency, error)
	GetByCode(ctx context.Context, code string) (*Currency, error)
}
