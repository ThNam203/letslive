package domains

import (
	"context"
	"sen1or/letslive/user/response"
	"time"

	"github.com/gofrs/uuid/v5"
)

type Follower struct {
	UserID     uuid.UUID `json:"userId" db:"user_id"`
	FollowerID uuid.UUID `json:"followerId" db:"follower_id"`
	FollowedAt time.Time `json:"createdAt" db:"created_at"`
}

type FollowRepository interface {
	FollowUser(ctx context.Context, followUser, followedUser uuid.UUID) *response.Response[any]
	UnfollowUser(ctx context.Context, followUser, followedUser uuid.UUID) *response.Response[any]
	GetFollowedUserIds(ctx context.Context, followerId uuid.UUID) ([]uuid.UUID, *response.Response[any])
}
