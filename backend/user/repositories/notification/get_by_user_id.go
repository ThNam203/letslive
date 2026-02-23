package notification

import (
	"context"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/pkg/logger"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r postgresNotificationRepo) GetByUserId(ctx context.Context, userId uuid.UUID, page int, pageSize int) ([]domains.Notification, int, *response.Response[any]) {
	offset := page * pageSize

	// get total count
	var total int
	err := r.dbConn.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1`, userId).Scan(&total)
	if err != nil {
		logger.Errorf(ctx, "failed to count notifications: %s", err)
		return nil, 0, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil, nil, nil,
		)
	}

	rows, err := r.dbConn.Query(ctx, `
		SELECT id, user_id, type, title, message, action_url, action_label, reference_id, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userId, pageSize, offset)
	if err != nil {
		logger.Errorf(ctx, "failed to query notifications: %s", err)
		return nil, 0, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil, nil, nil,
		)
	}
	defer rows.Close()

	notifications, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.Notification])
	if err != nil {
		logger.Errorf(ctx, "failed to scan notifications: %s", err)
		return nil, 0, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil, nil, nil,
		)
	}

	return notifications, total, nil
}
