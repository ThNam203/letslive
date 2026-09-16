package transaction

import (
	"context"
	"errors"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r postgresTransactionRepo) GetById(ctx context.Context, id uuid.UUID) (*domains.Transaction, error) {
	query := `
        select id, type, reference, status, actor_id, metadata, created_at
        from transactions
        where id = $1
    `
	rows, err := r.dbConn.Query(ctx, query, id)
	if err != nil {
		logger.Errorf(ctx, "db query error [gettransactionbyid: %v]", err)
		return nil, domains.ErrDatabaseQuery
	}

	tx, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Transaction])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrTransactionFailed
		}
		logger.Errorf(ctx, "db scan error [gettransactionbyid: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}
	return &tx, nil
}
