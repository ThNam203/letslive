package follower

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
)

func (r postgresFollowRepo) UnfollowUser(ctx context.Context, followUser, followedUser uuid.UUID) error {
	result, err := r.dbConn.Exec(ctx, `
		DELETE FROM followers
		WHERE user_id = $1 AND follower_id = $2
	`, followedUser, followUser)
	if err != nil {
		logger.Errorf(ctx, "failed to exec unfollow user: %s", err)
		return domains.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		logger.Errorf(ctx, "failed to unfollow user: %s", err)
		return domains.ErrDatabaseIssue
	}

	return nil
}
