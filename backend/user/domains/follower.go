package domains

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type Follower struct {
	UserID     uuid.UUID `json:"userId" db:"user_id"`
	FollowerID uuid.UUID `json:"followerId" db:"follower_id"`
	FollowedAt time.Time `json:"createdAt" db:"created_at"`
}
