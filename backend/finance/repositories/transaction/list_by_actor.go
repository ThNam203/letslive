package transaction

import (
	"context"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r postgresTransactionRepo) ListByActor(ctx context.Context, actorId uuid.UUID, page int, limit int) ([]domains.Transaction, int, *response.Response[any]) {
	countQuery := `select count(*) from transactions where actor_id = $1`
	var total int
	if err := r.dbConn.QueryRow(ctx, countQuery, actorId).Scan(&total); err != nil {
		logger.Errorf(ctx, "db count error [listtransactionsbyactor: %v]", err)
		return nil, 0, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	query := `
        select id, type, reference, status, actor_id, metadata, created_at
        from transactions
        where actor_id = $1
        order by created_at desc
        limit $2 offset $3
    `
	offset := page * limit
	rows, err := r.dbConn.Query(ctx, query, actorId, limit, offset)
	if err != nil {
		logger.Errorf(ctx, "db query error [listtransactionsbyactor: %v]", err)
		return nil, 0, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	txs, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.Transaction])
	if err != nil {
		logger.Errorf(ctx, "db scan error [listtransactionsbyactor: %v]", err)
		return nil, 0, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return txs, total, nil
}
