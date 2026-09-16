package notification

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
)

func (r postgresNotificationRepo) GetUnreadCount(ctx context.Context, userId uuid.UUID) (int, error) {
	var count int
	err := r.dbConn.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`, userId).Scan(&count)
	if err != nil {
		logger.Errorf(ctx, "failed to get unread notification count: %s", err)
		return 0, domains.ErrDatabaseQuery
	}

	return count, nil
}
