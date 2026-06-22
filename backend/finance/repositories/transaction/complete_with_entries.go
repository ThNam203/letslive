package transaction

import (
	"context"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

// CompleteWithEntries inserts ledger entries, updates account_balances cache, and transitions the
// transaction to 'completed' inside a single DB transaction. The DB zero-sum trigger on
// transactions.status validates sum(entries.amount) = 0 on the status transition; if it fails,
// the whole DB transaction is rolled back.
func (r postgresTransactionRepo) CompleteWithEntries(ctx context.Context, transactionId uuid.UUID, entries []domains.LedgerEntryDraft) *response.Response[any] {
	if len(entries) == 0 {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		)
	}

	dbTx, err := r.dbConn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Errorf(ctx, "db begin error [completewithentries: %v]", err)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	defer dbTx.Rollback(ctx)

	insertEntry := `
        insert into ledger_entries (transaction_id, account_id, currency_code, amount)
        values ($1, $2, $3, $4)
        returning id
    `
	upsertBalance := `
        insert into account_balances (account_id, currency_code, balance, last_entry_id, updated_at)
        values ($1, $2, $3, $4, current_timestamp)
        on conflict (account_id, currency_code) do update
          set balance = account_balances.balance + excluded.balance,
              last_entry_id = excluded.last_entry_id,
              updated_at = current_timestamp
    `

	for _, e := range entries {
		var entryId uuid.UUID
		if err := dbTx.QueryRow(ctx, insertEntry, transactionId, e.AccountId, e.CurrencyCode, e.Amount).Scan(&entryId); err != nil {
			logger.Errorf(ctx, "db insert ledger entry error [completewithentries: %v]", err)
			return response.NewResponseFromTemplate[any](
				response.RES_ERR_TRANSACTION_FAILED,
				nil,
				nil,
				nil,
			)
		}
		if _, err := dbTx.Exec(ctx, upsertBalance, e.AccountId, e.CurrencyCode, e.Amount, entryId); err != nil {
			logger.Errorf(ctx, "db upsert balance error [completewithentries: %v]", err)
			return response.NewResponseFromTemplate[any](
				response.RES_ERR_TRANSACTION_FAILED,
				nil,
				nil,
				nil,
			)
		}
	}

	// Transition transaction -> completed. The zero-sum trigger validates the ledger here.
	if _, err := dbTx.Exec(ctx, `update transactions set status = 'completed' where id = $1`, transactionId); err != nil {
		logger.Errorf(ctx, "db update transaction status error [completewithentries: %v]", err)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_TRANSACTION_FAILED,
			nil,
			nil,
			nil,
		)
	}

	if err := dbTx.Commit(ctx); err != nil {
		logger.Errorf(ctx, "db commit error [completewithentries: %v]", err)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return nil
}
