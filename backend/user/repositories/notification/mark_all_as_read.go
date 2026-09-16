package notification

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
)

func (r postgresNotificationRepo) MarkAllAsRead(ctx context.Context, userId uuid.UUID) error {
	_, err := r.dbConn.Exec(ctx, `
		UPDATE notifications SET is_read = true WHERE user_id = $1 AND is_read = false
	`, userId)
	if err != nil {
		logger.Errorf(ctx, "failed to mark all notifications as read: %s", err)
		return domains.ErrDatabaseQuery
	}

	return nil
}
