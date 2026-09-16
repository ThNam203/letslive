package currency

import (
	"context"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/jackc/pgx/v5"
)

func (r postgresCurrencyRepo) List(ctx context.Context) ([]domains.Currency, error) {
	query := `
        select code, name, precision
        from currencies
        order by code
    `
	rows, err := r.dbConn.Query(ctx, query)
	if err != nil {
		logger.Errorf(ctx, "db query error [listcurrencies: %v]", err)
		return nil, domains.ErrDatabaseQuery
	}

	currencies, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.Currency])
	if err != nil {
		logger.Errorf(ctx, "db scan error [listcurrencies: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}
	return currencies, nil
}
