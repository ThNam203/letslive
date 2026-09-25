package transaction

import (
	"context"
	"errors"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

// CompleteWithEntries inserts ledger entries, updates the account_balances cache, and
// transitions the transaction to 'completed' inside a single DB transaction.
//
// The transaction row is locked FOR UPDATE first, which serializes concurrent
// completions (e.g. duplicate webhook deliveries): the second caller blocks until the
// first commits, then sees status 'completed' and returns success without inserting
// anything. User wallets are rejected if an entry would drive their balance negative,
// which also closes the check-then-debit race between concurrent purchases. The DB
// zero-sum trigger validates sum(entries.amount) = 0 on the status transition.
func (r postgresTransactionRepo) CompleteWithEntries(ctx context.Context, transactionId uuid.UUID, entries []domains.LedgerEntryDraft) error {
	if len(entries) == 0 {
		return domains.ErrInvalidInput
	}

	dbTx, err := r.dbConn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Errorf(ctx, "db begin error [completewithentries: %v]", err)
		return domains.ErrDatabaseIssue
	}
	defer dbTx.Rollback(ctx)

	var status domains.ProcessStatus
	err = dbTx.QueryRow(ctx, `select status from transactions where id = $1 for update`, transactionId).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domains.ErrTransactionNotFound
		}
		logger.Errorf(ctx, "db lock transaction error [completewithentries: %v]", err)
		return domains.ErrDatabaseIssue
	}

	switch status {
	case domains.ProcessStatusCompleted:
		// already completed (duplicate webhook / retry): idempotent success
		return nil
	case domains.ProcessStatusCreated, domains.ProcessStatusProcessing:
		// completable
	default:
		logger.Errorf(ctx, "cannot complete transaction %s in status %s [completewithentries]", transactionId, status)
		return domains.ErrTransactionFailed
	}

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
        returning balance
    `

	for _, e := range entries {
		var entryId uuid.UUID
		if err := dbTx.QueryRow(ctx, insertEntry, transactionId, e.AccountId, e.CurrencyCode, e.Amount).Scan(&entryId); err != nil {
			logger.Errorf(ctx, "db insert ledger entry error [completewithentries: %v]", err)
			return domains.ErrTransactionFailed
		}

		var newBalance int64
		if err := dbTx.QueryRow(ctx, upsertBalance, e.AccountId, e.CurrencyCode, e.Amount, entryId).Scan(&newBalance); err != nil {
			logger.Errorf(ctx, "db upsert balance error [completewithentries: %v]", err)
			return domains.ErrTransactionFailed
		}

		// user wallets may not be overdrawn; platform-side accounts (escrow) may go
		// negative by design as the mint counter-account
		if newBalance < 0 {
			var accountType domains.AccountType
			if err := dbTx.QueryRow(ctx, `select type from accounts where id = $1`, e.AccountId).Scan(&accountType); err != nil {
				logger.Errorf(ctx, "db select account type error [completewithentries: %v]", err)
				return domains.ErrTransactionFailed
			}
			if accountType == domains.AccountTypeUserWallet {
				return domains.ErrInsufficientBalance
			}
		}
	}

	// Transition transaction -> completed. The zero-sum trigger validates the ledger here.
	if _, err := dbTx.Exec(ctx, `update transactions set status = 'completed' where id = $1`, transactionId); err != nil {
		logger.Errorf(ctx, "db update transaction status error [completewithentries: %v]", err)
		return domains.ErrTransactionFailed
	}

	if err := dbTx.Commit(ctx); err != nil {
		logger.Errorf(ctx, "db commit error [completewithentries: %v]", err)
		return domains.ErrDatabaseIssue
	}
	return nil
}
