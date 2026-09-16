package notification

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/domains"

	"github.com/jackc/pgx/v5"
)

func (r postgresNotificationRepo) Create(ctx context.Context, n domains.Notification) (*domains.Notification, error) {
	rows, err := r.dbConn.Query(ctx, `
		INSERT INTO notifications (user_id, type, title, message, action_url, action_label, reference_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, type, title, message, action_url, action_label, reference_id, is_read, created_at
	`, n.UserId, n.Type, n.Title, n.Message, n.ActionUrl, n.ActionLabel, n.ReferenceId)
	if err != nil {
		logger.Errorf(ctx, "failed to insert notification: %s", err)
		return nil, domains.ErrDatabaseQuery
	}
	defer rows.Close()

	created, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Notification])
	if err != nil {
		logger.Errorf(ctx, "failed to scan created notification: %s", err)
		return nil, domains.ErrDatabaseQuery
	}

	return &created, nil
}
