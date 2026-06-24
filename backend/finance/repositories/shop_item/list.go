package shopitemrepo

import (
	"context"
	"errors"

	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/response"

	"github.com/jackc/pgx/v5"
)

func (r *postgresShopItemRepo) List(ctx context.Context) ([]domains.ShopItem, *response.Response[any]) {
	query := `
		SELECT id, name, description, image_url, animation_url, price, is_active, created_at
		FROM shop_items
		WHERE is_active = true
		ORDER BY created_at ASC
	`

	rows, err := r.dbConn.Query(ctx, query)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []domains.ShopItem{}, nil
		}
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_QUERY, nil, nil, nil)
	}

	items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.ShopItem])
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_ISSUE, nil, nil, nil)
	}

	return items, nil
}
