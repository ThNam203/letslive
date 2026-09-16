package currency

import (
	"context"
	"errors"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/jackc/pgx/v5"
)

func (r postgresCurrencyRepo) GetByCode(ctx context.Context, code string) (*domains.Currency, error) {
	query := `
        select code, name, precision
        from currencies
        where code = $1
    `
	rows, err := r.dbConn.Query(ctx, query, code)
	if err != nil {
		logger.Errorf(ctx, "db query error [getcurrencybycode: %v]", err)
		return nil, domains.ErrDatabaseQuery
	}

	currency, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Currency])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrUnsupportedCurrency
		}
		logger.Errorf(ctx, "db scan error [getcurrencybycode: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}
	return &currency, nil
}
