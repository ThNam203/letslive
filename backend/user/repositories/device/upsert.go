package devicerepo

import (
	"context"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/domains"

	"github.com/jackc/pgx/v5"
)

func (r *postgresDeviceRepo) Upsert(ctx context.Context, token domains.DeviceToken) (*domains.DeviceToken, error) {
	rows, err := r.dbConn.Query(ctx, `
		INSERT INTO device_tokens (user_id, token, platform, locale)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (token) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			platform = EXCLUDED.platform,
			locale = EXCLUDED.locale,
			last_seen_at = clock_timestamp()
		RETURNING id, user_id, token, platform, locale, created_at, last_seen_at
	`, token.UserId, token.Token, token.Platform, token.Locale)
	if err != nil {
		logger.Errorf(ctx, "failed to upsert device token: %s", err)
		return nil, domains.ErrDatabaseQuery
	}

	saved, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[domains.DeviceToken])
	if err != nil {
		logger.Errorf(ctx, "failed to upsert/scan device token: %s", err)
		return nil, domains.ErrDatabaseQuery
	}

	return &saved, nil
}
