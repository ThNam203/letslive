package currency

import (
	"context"
	"errors"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/jackc/pgx/v5"
)

func (r postgresCurrencyRepo) GetByCode(ctx context.Context, code string) (*domains.Currency, *response.Response[any]) {
	query := `
        select code, name, precision
        from currencies
        where code = $1
    `
	rows, err := r.dbConn.Query(ctx, query, code)
	if err != nil {
		logger.Errorf(ctx, "db query error [getcurrencybycode: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	currency, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Currency])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_UNSUPPORTED_CURRENCY,
				nil,
				nil,
				nil,
			)
		}
		logger.Errorf(ctx, "db scan error [getcurrencybycode: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return &currency, nil
}
