package inventoryrepo

import (
	"context"
	"errors"

	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresInventoryRepo) Deduct(ctx context.Context, userID, shopItemID uuid.UUID) (*domains.UserInventory, error) {
	query := `
		UPDATE user_inventory
		SET quantity = quantity - 1, updated_at = now()
		WHERE user_id = $1 AND shop_item_id = $2 AND quantity >= 1
		RETURNING id, user_id, shop_item_id, quantity, updated_at
	`

	rows, err := r.dbConn.Query(ctx, query, userID, shopItemID)
	if err != nil {
		return nil, domains.ErrDatabaseQuery
	}

	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.UserInventory])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrInsufficientInventory
		}
		return nil, domains.ErrDatabaseIssue
	}

	return &item, nil
}
