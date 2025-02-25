package dto

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type CreateUserRequestDTO struct {
	Username   string `json:"username" validate:"required,gte=6,lte=20"`
	Email      string `json:"email" validate:"required,email"`
	IsVerified bool   `json:"isVerified,omitempty" validate:"omitempty"`
}

type GetUserRequestDTO struct{}

type GetUserResponseDTO struct {
	Id                uuid.UUID `json:"id"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	IsOnline          bool      `json:"isOnline"`
	IsVerified        bool      `json:"isVerified"`
	CreatedAt         time.Time `json:"createdAt"`
	PhoneNumber       *string   `json:"phoneNumber,omitempty"`
	Bio               *string   `json:"bio,omitempty"`
	DisplayName       *string   `json:"displayName,omitempty"`
	ProfilePicture    *string   `json:"profilePicture,omitempty"`
	BackgroundPicture *string   `json:"backgroundPicture,omitempty"`

	VODs []GetLivestreamResponseDTO `json:"vods"`
}

type GetUserByStreamAPIKeyRequestDTO struct{}

type UpdateUserRequestDTO struct {
	Id          uuid.UUID `json:"id" validate:"uuid"`
	Username    *string   `json:"username,omitempty" validate:"omitempty,gte=6,lte=20"`
	IsOnline    *bool     `json:"isOnline,omitempty" validate:""`
	IsActive    *bool     `json:"isActive,omitempty"`
	PhoneNumber *string   `json:"phoneNumber,omitempty"`
	Bio         *string   `json:"bio,omitempty"`
	DisplayName *string   `json:"displayName,omitempty" validate:"omitempty,gte=6,lte=20"`
}
