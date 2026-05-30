package payment

import (
	"context"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r postgresPaymentRepo) ListByActor(ctx context.Context, actorId uuid.UUID, page int, limit int) ([]domains.Payment, int, *response.Response[any]) {
	countQuery := `
        select count(*)
        from payments p
        join transactions t on t.id = p.transaction_id
        where t.actor_id = $1
    `
	var total int
	if err := r.dbConn.QueryRow(ctx, countQuery, actorId).Scan(&total); err != nil {
		logger.Errorf(ctx, "db count error [listpaymentsbyactor: %v]", err)
		return nil, 0, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	query := `
        select p.id, p.provider, p.provider_ref, p.currency_code, p.amount, p.status, p.transaction_id, p.created_at
        from payments p
        join transactions t on t.id = p.transaction_id
        where t.actor_id = $1
        order by p.created_at desc
        limit $2 offset $3
    `
	offset := page * limit
	rows, err := r.dbConn.Query(ctx, query, actorId, limit, offset)
	if err != nil {
		logger.Errorf(ctx, "db query error [listpaymentsbyactor: %v]", err)
		return nil, 0, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	payments, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.Payment])
	if err != nil {
		logger.Errorf(ctx, "db scan error [listpaymentsbyactor: %v]", err)
		return nil, 0, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return payments, total, nil
}
