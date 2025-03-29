package dto

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type CreateUserRequestDTO struct {
	Username     string `json:"username" validate:"required,gte=4,lte=50"`
	Email        string `json:"email" validate:"required,email"`
	IsVerified   bool   `json:"isVerified,omitempty" validate:"omitempty"`
	AuthProvider string `json:"authProvider" validate:"oneof=google local"`
}

type GetUserRequestDTO struct{}

type GetUserPublicResponseDTO struct {
	Id                uuid.UUID `json:"id"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	IsVerified        bool      `json:"isVerified"`
	CreatedAt         time.Time `json:"createdAt"`
	PhoneNumber       *string   `json:"phoneNumber,omitempty"`
	Bio               *string   `json:"bio,omitempty"`
	DisplayName       *string   `json:"displayName,omitempty"`
	ProfilePicture    *string   `json:"profilePicture,omitempty"`
	BackgroundPicture *string   `json:"backgroundPicture,omitempty"`
	FollowerCount     *int      `json:"followerCount,omitempty"`

	// whether or not the current fetching user is following the fetched user
	IsFollowing *bool `json:"isFollowing,omitempty"`

	LivestreamInformation `json:"livestreamInformation"`
}

type LivestreamInformation struct {
	UserID       uuid.UUID `db:"user_id,omitempty" json:"userId"`
	Title        *string   `db:"title,omitempty" json:"title"`
	Description  *string   `db:"description,omitempty" json:"description"`
	ThumbnailURL *string   `db:"thumbnail_url,omitempty" json:"thumbnailUrl"`
}

type GetUserByStreamAPIKeyRequestDTO struct{}

type UpdateUserRequestDTO struct {
	Id          uuid.UUID `json:"id" validate:"uuid"`
	Username    *string   `json:"username,omitempty" validate:"omitempty,gte=6,lte=20"`
	IsActive    *bool     `json:"isActive,omitempty"`
	PhoneNumber *string   `json:"phoneNumber,omitempty"`
	Bio         *string   `json:"bio,omitempty"`
	DisplayName *string   `json:"displayName,omitempty" validate:"omitempty,gte=6,lte=20"`
}
