package inventoryrepo

import (
	"context"

	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresInventoryRepo) GetByUserId(ctx context.Context, userID uuid.UUID, page, limit int) ([]domains.UserInventory, int, *response.Response[any]) {
	var total int
	if err := r.dbConn.QueryRow(ctx, `SELECT count(*) FROM user_inventory WHERE user_id = $1 AND quantity > 0`, userID).Scan(&total); err != nil {
		return nil, 0, response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_QUERY, nil, nil, nil)
	}

	rows, err := r.dbConn.Query(ctx, `
		SELECT id, user_id, shop_item_id, quantity, updated_at
		FROM user_inventory
		WHERE user_id = $1 AND quantity > 0
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, page*limit)
	if err != nil {
		return nil, 0, response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_QUERY, nil, nil, nil)
	}

	items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.UserInventory])
	if err != nil {
		return nil, 0, response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_ISSUE, nil, nil, nil)
	}

	return items, total, nil
}
