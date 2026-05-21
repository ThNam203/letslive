package payment

import (
	"context"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/jackc/pgx/v5"
)

func (r postgresPaymentRepo) Create(ctx context.Context, p domains.Payment) (*domains.Payment, *response.Response[any]) {
	query := `
        insert into payments (provider, provider_ref, currency_code, amount, status, transaction_id)
        values ($1, $2, $3, $4, $5, $6)
        returning id, provider, provider_ref, currency_code, amount, status, transaction_id, created_at
    `
	rows, err := r.dbConn.Query(ctx, query, p.Provider, p.ProviderRef, p.CurrencyCode, p.Amount, p.Status, p.TransactionId)
	if err != nil {
		logger.Errorf(ctx, "db query error [createpayment: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	created, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Payment])
	if err != nil {
		logger.Errorf(ctx, "db scan error [createpayment: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_PAYMENT_FAILED,
			nil,
			nil,
			nil,
		)
	}
	return &created, nil
}
