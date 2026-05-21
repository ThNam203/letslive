package domains

import (
	"context"
	response "sen1or/letslive/finance/response"
)

type Currency struct {
	Code      string `json:"code" db:"code"`
	Name      string `json:"name" db:"name"`
	Precision int    `json:"precision" db:"precision"`
}

type CurrencyRepository interface {
	List(ctx context.Context) ([]Currency, *response.Response[any])
	GetByCode(ctx context.Context, code string) (*Currency, *response.Response[any])
}
