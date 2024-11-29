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
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	IsOnline     bool      `json:"isOnline"`
	CreatedAt    time.Time `json:"createdAt"`
	StreamAPIKey uuid.UUID `json:"streamAPIKey"`
}

type GetUserRequestDTO struct{}

type GetUserResponseDTO struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	IsOnline     bool      `json:"isOnline"`
	CreatedAt    time.Time `json:"createdAt"`
	StreamAPIKey uuid.UUID `json:"streamAPIKey"`
}

type GetUserByStreamAPIKeyRequestDTO struct{}

type GetUserByStreamAPIKeyResponseDTO struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	IsOnline     bool      `json:"isOnline"`
	CreatedAt    time.Time `json:"createdAt"`
	StreamAPIKey uuid.UUID `json:"streamAPIKey"`
}

type UpdateUserRequestDTO struct {
	ID       uuid.UUID `json:"id" validate:"uuid"`
	Username *string   `json:"username" validate:"gte=6,lte=20"`
	IsOnline *bool     `json:"isOnline" validate:""`
}

type UpdateUserResponseDTO struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	IsOnline     bool      `json:"isOnline"`
	StreamAPIKey uuid.UUID `json:"streamAPIKey"`
	CreatedAt    time.Time `json:"createdAt"`
}
