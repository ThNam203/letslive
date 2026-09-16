package notification

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
)

func (r postgresNotificationRepo) DeleteById(ctx context.Context, notificationId uuid.UUID, userId uuid.UUID) error {
	result, err := r.dbConn.Exec(ctx, `
		DELETE FROM notifications WHERE id = $1 AND user_id = $2
	`, notificationId, userId)
	if err != nil {
		logger.Errorf(ctx, "failed to delete notification: %s", err)
		return domains.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		return domains.ErrNotificationNotFound
	}

	return nil
}
