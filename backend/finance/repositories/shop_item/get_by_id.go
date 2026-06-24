package shopitemrepo

import (
	"context"
	"errors"

	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/response"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *postgresShopItemRepo) GetById(ctx context.Context, id uuid.UUID) (*domains.ShopItem, *response.Response[any]) {
	query := `
		SELECT id, name, description, image_url, animation_url, price, is_active, created_at
		FROM shop_items
		WHERE id = $1 AND is_active = true
	`

	rows, err := r.dbConn.Query(ctx, query, id)
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_QUERY, nil, nil, nil)
	}

	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.ShopItem])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](response.RES_ERR_SHOP_ITEM_NOT_FOUND, nil, nil, nil)
		}
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_ISSUE, nil, nil, nil)
	}

	return &item, nil
}
