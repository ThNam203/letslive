package transaction

import (
	"context"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/jackc/pgx/v5"
)

func (r postgresTransactionRepo) Create(ctx context.Context, tx domains.Transaction) (*domains.Transaction, *response.Response[any]) {
	query := `
        insert into transactions (type, reference, status, actor_id, metadata)
        values ($1, $2, $3, $4, $5)
        returning id, type, reference, status, actor_id, metadata, created_at
    `
	rows, err := r.dbConn.Query(ctx, query, tx.Type, tx.Reference, tx.Status, tx.ActorId, tx.Metadata)
	if err != nil {
		logger.Errorf(ctx, "db query error [createtransaction: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	created, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Transaction])
	if err != nil {
		logger.Errorf(ctx, "db scan error [createtransaction: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_TRANSACTION_FAILED,
			nil,
			nil,
			nil,
		)
	}
	return &created, nil
}
