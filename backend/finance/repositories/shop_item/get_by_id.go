package shopitemrepo

import (
	"context"
	"errors"

	"sen1or/letslive/finance/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresShopItemRepo) GetById(ctx context.Context, id uuid.UUID) (*domains.ShopItem, error) {
	query := `
		SELECT id, name, description, image_url, animation_url, price, currency_code, is_active, created_at
		FROM shop_items
		WHERE id = $1 AND is_active = true
	`

	rows, err := r.dbConn.Query(ctx, query, id)
	if err != nil {
		return nil, domains.ErrDatabaseQuery
	}

	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.ShopItem])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrShopItemNotFound
		}
		return nil, domains.ErrDatabaseIssue
	}

	return &item, nil
}
