package transaction

import (
	"context"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r postgresTransactionRepo) GetEntriesForAccount(ctx context.Context, transactionId uuid.UUID, accountId uuid.UUID) ([]domains.LedgerEntry, *response.Response[any]) {
	query := `
        select id, transaction_id, account_id, currency_code, amount, created_at
        from ledger_entries
        where transaction_id = $1 and account_id = $2
        order by created_at asc
    `
	rows, err := r.dbConn.Query(ctx, query, transactionId, accountId)
	if err != nil {
		logger.Errorf(ctx, "db query error [getentriesforaccount: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	entries, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.LedgerEntry])
	if err != nil {
		logger.Errorf(ctx, "db scan error [getentriesforaccount: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return entries, nil
}
