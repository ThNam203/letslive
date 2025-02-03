package dto

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type CreateUserRequestDTO struct {
	Username string `json:"username" validate:"required,gte=6,lte=20"`
	Email    string `json:"email" validate:"required,email"`
}

type CreateUserResponseDTO struct {
	Id           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	IsOnline     bool      `json:"isOnline"`
	CreatedAt    time.Time `json:"createdAt"`
	StreamAPIKey uuid.UUID `json:"streamAPIKey"`
}

type GetUserRequestDTO struct{}

type GetUserResponseDTO struct {
	Id           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	IsOnline     bool      `json:"isOnline"`
	CreatedAt    time.Time `json:"createdAt"`
	StreamAPIKey uuid.UUID `json:"streamAPIKey"`
	VODs         []string  `json:"vods"`
}

type GetUserByStreamAPIKeyRequestDTO struct{}

type GetUserByStreamAPIKeyResponseDTO struct {
	Id           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	IsOnline     bool      `json:"isOnline"`
	CreatedAt    time.Time `json:"createdAt"`
	StreamAPIKey uuid.UUID `json:"streamAPIKey"`
}

type UpdateUserRequestDTO struct {
	Id       uuid.UUID `json:"id" validate:"uuid"`
	Username *string   `json:"username,omitempty" validate:"omitempty,gte=6,lte=20"`
	IsOnline *bool     `json:"isOnline,omitempty" validate:""`
}

type UpdateUserResponseDTO struct {
	Id           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	IsOnline     bool      `json:"isOnline"`
	StreamAPIKey uuid.UUID `json:"streamAPIKey"`
	CreatedAt    time.Time `json:"createdAt"`
}
