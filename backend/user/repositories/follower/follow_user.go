package follower

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
)

func (r postgresFollowRepo) FollowUser(ctx context.Context, followUser, followedUser uuid.UUID) error {
	result, err := r.dbConn.Exec(ctx, `
		INSERT INTO followers (user_id, follower_id)
		VALUES ($1, $2)
	`, followedUser, followUser)
	if err != nil {
		logger.Errorf(ctx, "failed to exec follow user: %s", err)
		return domains.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		logger.Errorf(ctx, "failed to follow user: %s", err)
		return domains.ErrDatabaseIssue
	}

	return nil
}
