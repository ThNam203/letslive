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

type UpdateUserRequestDTO struct {
	ID       uuid.UUID `json:"user" validate:"required,uuid"`
	Username string    `json:"username" validate:"required,gte=6,lte=20"`
}

type UpdateUserResponseDTO struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	IsOnline     bool      `json:"isOnline"`
	CreatedAt    time.Time `json:"createdAt"`
	StreamAPIKey uuid.UUID `json:"streamAPIKey"`
}
