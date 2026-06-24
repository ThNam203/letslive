package giftrepo

import (
	"context"

	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresGiftRepo) ListBySender(ctx context.Context, senderID uuid.UUID, page, limit int) ([]domains.Gift, int, *response.Response[any]) {
	var total int
	if err := r.dbConn.QueryRow(ctx, `SELECT count(*) FROM gifts WHERE sender_user_id = $1`, senderID).Scan(&total); err != nil {
		return nil, 0, response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_QUERY, nil, nil, nil)
	}

	rows, err := r.dbConn.Query(ctx, `
		SELECT id, sender_user_id, recipient_user_id, shop_item_id, quantity, message, sent_at
		FROM gifts WHERE sender_user_id = $1
		ORDER BY sent_at DESC LIMIT $2 OFFSET $3
	`, senderID, limit, page*limit)
	if err != nil {
		return nil, 0, response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_QUERY, nil, nil, nil)
	}

	gifts, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.Gift])
	if err != nil {
		return nil, 0, response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_ISSUE, nil, nil, nil)
	}

	return gifts, total, nil
}
