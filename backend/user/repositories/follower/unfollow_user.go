package follower

import (
	"context"
	"sen1or/letslive/user/pkg/logger"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

func (r postgresFollowRepo) UnfollowUser(ctx context.Context, followUser, followedUser uuid.UUID) *response.Response[any] {
	result, err := r.dbConn.Exec(ctx, `
		DELETE FROM followers
		WHERE user_id = $1 AND follower_id = $2
	`, followedUser, followUser)
	if err != nil {
		logger.Errorf(ctx, "failed to exec unfollow user: %s", err)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	if result.RowsAffected() == 0 {
		logger.Errorf(ctx, "failed to unfollow user: %s", err)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return nil
}
