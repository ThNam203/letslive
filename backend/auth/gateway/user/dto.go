package user

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type CreateUserRequestDTO struct {
	Username   string `json:"username" validate:"required,gte=6,lte=20"`
	Email      string `json:"email" validate:"required,email"`
	IsVerified bool   `json:"isVerified" validate:"required,email"`
}

type CreateUserResponseDTO struct {
	Id                uuid.UUID `json:"id"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	IsVerified        bool      `json:"isVerified"`
	IsOnline          bool      `json:"isOnline"`
	IsActive          bool      `json:"isActive"`
	CreatedAt         time.Time `json:"createdAt"`
	StreamAPIKey      uuid.UUID `json:"streamAPIKey"`
	DisplayName       *string   `json:"displayName,omitempty"`
	PhoneNumber       *string   `json:"phoneNumber,omitempty"`
	Bio               *string   `json:"bio,omitempty"`
	ProfilePicture    *string   `json:"profilePicture,omitempty"`
	BackgroundPicture *string   `json:"backgroundPicture,omitempty"`
}
