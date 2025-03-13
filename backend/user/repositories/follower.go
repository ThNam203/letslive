package repositories

import (
	"context"
	servererrors "sen1or/letslive/user/errors"
	"sen1or/letslive/user/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FollowRepository interface {
	FollowUser(followUser, followedUser uuid.UUID) *servererrors.ServerError
	UnfollowUser(followUser, followedUser uuid.UUID) *servererrors.ServerError
}

type postgresFollowRepo struct {
	dbConn *pgxpool.Pool
}

func NewFollowRepository(conn *pgxpool.Pool) FollowRepository {
	return &postgresFollowRepo{
		dbConn: conn,
	}
}

func (r postgresFollowRepo) FollowUser(followUser, followedUser uuid.UUID) *servererrors.ServerError {
	result, err := r.dbConn.Exec(context.Background(), `
		INSERT INTO followers (user_id, follower_id)
		VALUES ($1, $2)
	`, followedUser, followUser)
	if err != nil {
		logger.Errorf("failed to exec follow user: %s", err)
		return servererrors.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		logger.Errorf("failed to follow user: %s", err)
		return servererrors.ErrDatabaseIssue
	}

	return nil
}
func (r postgresFollowRepo) UnfollowUser(followUser, followedUser uuid.UUID) *servererrors.ServerError {
	result, err := r.dbConn.Exec(context.Background(), `
		DELETE FROM followers
		WHERE user_id = $1 AND follower_id = $2
	`, followedUser, followUser)
	if err != nil {
		logger.Errorf("failed to exec unfollow user: %s", err)
		return servererrors.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		logger.Errorf("failed to unfollow user: %s", err)
		return servererrors.ErrDatabaseIssue
	}

	return nil
}
