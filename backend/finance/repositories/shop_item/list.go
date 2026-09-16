package shopitemrepo

import (
	"context"
	"errors"

	"sen1or/letslive/finance/domains"

	"github.com/jackc/pgx/v5"
)

func (r *postgresShopItemRepo) List(ctx context.Context) ([]domains.ShopItem, error) {
	query := `
		SELECT id, name, description, image_url, animation_url, price, currency_code, is_active, created_at
		FROM shop_items
		WHERE is_active = true
		ORDER BY created_at ASC
	`

	rows, err := r.dbConn.Query(ctx, query)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []domains.ShopItem{}, nil
		}
		return nil, domains.ErrDatabaseQuery
	}

	items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.ShopItem])
	if err != nil {
		return nil, domains.ErrDatabaseIssue
	}

	return items, nil
}
