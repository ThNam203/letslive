package follower

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r postgresFollowRepo) GetFollowedUserIds(ctx context.Context, followerId uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT user_id FROM followers WHERE follower_id = $1
	`, followerId)
	if err != nil {
		logger.Errorf(ctx, "failed to get followed user ids: %s", err)
		return nil, domains.ErrDatabaseQuery
	}

	ids, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (uuid.UUID, error) {
		var id uuid.UUID
		err := row.Scan(&id)
		return id, err
	})
	if err != nil {
		logger.Errorf(ctx, "failed to collect followed user ids: %s", err)
		return nil, domains.ErrDatabaseIssue
	}

	return ids, nil
}
