package inventoryrepo

import (
	"context"

	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresInventoryRepo) Upsert(ctx context.Context, userID, shopItemID uuid.UUID, quantityToAdd int) (*domains.UserInventory, *response.Response[any]) {
	query := `
		INSERT INTO user_inventory (user_id, shop_item_id, quantity, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (user_id, shop_item_id)
		DO UPDATE SET quantity = user_inventory.quantity + $3, updated_at = now()
		RETURNING id, user_id, shop_item_id, quantity, updated_at
	`

	rows, err := r.dbConn.Query(ctx, query, userID, shopItemID, quantityToAdd)
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_QUERY, nil, nil, nil)
	}

	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.UserInventory])
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_ISSUE, nil, nil, nil)
	}

	return &item, nil
}
