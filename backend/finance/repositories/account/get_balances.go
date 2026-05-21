package account

import (
	"context"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r postgresAccountRepo) GetBalances(ctx context.Context, accountId uuid.UUID) ([]domains.AccountBalance, *response.Response[any]) {
	query := `
        select account_id, currency_code, balance, last_entry_id, updated_at
        from account_balances
        where account_id = $1
    `
	rows, err := r.dbConn.Query(ctx, query, accountId)
	if err != nil {
		logger.Errorf(ctx, "db query error [getbalances: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	balances, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.AccountBalance])
	if err != nil {
		logger.Errorf(ctx, "db scan error [getbalances: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return balances, nil
}
