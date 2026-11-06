package devicerepo

import (
	"context"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresDeviceRepo) ListTokensByUserID(ctx context.Context, userID uuid.UUID) ([]domains.DeviceToken, error) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT id, user_id, token, platform, locale, created_at, last_seen_at
		FROM device_tokens
		WHERE user_id = $1
		ORDER BY last_seen_at DESC
	`, userID)
	if err != nil {
		logger.Errorf(ctx, "failed to query device tokens: %s", err)
		return nil, domains.ErrDatabaseQuery
	}

	tokens, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.DeviceToken])
	if err != nil {
		logger.Errorf(ctx, "failed to scan device tokens: %s", err)
		return nil, domains.ErrDatabaseQuery
	}

	return tokens, nil
}
