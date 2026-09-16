package payment

import (
	"context"
	"errors"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/jackc/pgx/v5"
)

func (r postgresPaymentRepo) GetByProviderRef(ctx context.Context, provider domains.PaymentProvider, providerRef string) (*domains.Payment, error) {
	query := `
        select id, provider, provider_ref, currency_code, amount, status, transaction_id, created_at
        from payments
        where provider = $1 and provider_ref = $2
    `
	rows, err := r.dbConn.Query(ctx, query, provider, providerRef)
	if err != nil {
		logger.Errorf(ctx, "db query error [getpaymentbyproviderref: %v]", err)
		return nil, domains.ErrDatabaseQuery
	}

	p, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Payment])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrPaymentNotFound
		}
		logger.Errorf(ctx, "db scan error [getpaymentbyproviderref: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}
	return &p, nil
}
