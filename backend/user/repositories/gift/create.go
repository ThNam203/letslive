package giftrepo

import (
	"context"

	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/response"

	"github.com/jackc/pgx/v5"
)

func (r *postgresGiftRepo) Create(ctx context.Context, gift domains.Gift) (*domains.Gift, *response.Response[any]) {
	rows, err := r.dbConn.Query(ctx, `
		INSERT INTO gifts (sender_user_id, recipient_user_id, shop_item_id, quantity, message)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, sender_user_id, recipient_user_id, shop_item_id, quantity, message, sent_at
	`, gift.SenderUserId, gift.RecipientUserId, gift.ShopItemId, gift.Quantity, gift.Message)
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_QUERY, nil, nil, nil)
	}

	created, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Gift])
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_ISSUE, nil, nil, nil)
	}

	return &created, nil
}
