package dto

import (
	"sen1or/lets-live/user/domains"
	livestreamgateway "sen1or/lets-live/user/gateway/livestream"
	"time"

	"github.com/gofrs/uuid/v5"
)

type CreateUserRequestDTO struct {
	Username     string `json:"username" validate:"required,gte=6,lte=20"`
	Email        string `json:"email" validate:"required,email"`
	IsVerified   bool   `json:"isVerified,omitempty" validate:"omitempty"`
	AuthProvider string `json:"authProvider" validate:"oneof=google local"`
}

type GetUserRequestDTO struct{}

type GetUserResponseDTO struct {
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
	IsFollowing       *bool     `json:"isFollowing,omitempty"`

	IsLivestreaming bool `json:"isLivestreaming"`

	domains.LivestreamInformation `json:"livestreamInformation"`
	VODs                          []livestreamgateway.GetLivestreamResponseDTO `json:"vods,omitempty"`
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
