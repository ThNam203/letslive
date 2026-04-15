package events

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// User event types.
const (
	UserCreated    = "user.created"
	UserUpdated    = "user.updated"
	UserFollowed   = "user.followed"
	UserUnfollowed = "user.unfollowed"
)

// UserCreatedEvent is emitted when a new user is registered.
type UserCreatedEvent struct {
	UserId    uuid.UUID `json:"userId"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

// UserUpdatedEvent is emitted when a user profile is updated.
type UserUpdatedEvent struct {
	UserId      uuid.UUID `json:"userId"`
	Username    *string   `json:"username,omitempty"`
	DisplayName *string   `json:"displayName,omitempty"`
}

// UserFollowedEvent is emitted when a user follows another user.
type UserFollowedEvent struct {
	UserId     uuid.UUID `json:"userId"`
	FollowerId uuid.UUID `json:"followerId"`
	FollowedAt time.Time `json:"followedAt"`
}

// UserUnfollowedEvent is emitted when a user unfollows another user.
type UserUnfollowedEvent struct {
	UserId     uuid.UUID `json:"userId"`
	FollowerId uuid.UUID `json:"followerId"`
}
