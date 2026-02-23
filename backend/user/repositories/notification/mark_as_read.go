package notification

import (
	"context"
	"sen1or/letslive/user/pkg/logger"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

func (r postgresNotificationRepo) MarkAsRead(ctx context.Context, notificationId uuid.UUID, userId uuid.UUID) *response.Response[any] {
	result, err := r.dbConn.Exec(ctx, `
		UPDATE notifications SET is_read = true WHERE id = $1 AND user_id = $2
	`, notificationId, userId)
	if err != nil {
		logger.Errorf(ctx, "failed to mark notification as read: %s", err)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil, nil, nil,
		)
	}

	if result.RowsAffected() == 0 {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_NOTIFICATION_NOT_FOUND,
			nil, nil, nil,
		)
	}

	return nil
}
