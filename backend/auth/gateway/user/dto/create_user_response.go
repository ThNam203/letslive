package dto

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type CreateUserResponseDTO struct {
	Id           uuid.UUID `json:"id"`
	Username     *string   `json:"username,omitempty"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"createdAt"`
	StreamAPIKey uuid.UUID `json:"streamAPIKey"`
	PhoneNumber  *string   `json:"phoneNumber,omitempty"`
}

