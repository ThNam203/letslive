package notification

import (
	"context"
	"sen1or/letslive/user/pkg/logger"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

func (r postgresNotificationRepo) MarkAllAsRead(ctx context.Context, userId uuid.UUID) *response.Response[any] {
	_, err := r.dbConn.Exec(ctx, `
		UPDATE notifications SET is_read = true WHERE user_id = $1 AND is_read = false
	`, userId)
	if err != nil {
		logger.Errorf(ctx, "failed to mark all notifications as read: %s", err)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil, nil, nil,
		)
	}

	return nil
}
