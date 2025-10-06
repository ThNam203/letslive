package user

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type CreateUserRequestDTO struct {
	Username     string       `json:"username"`
	Email        string       `json:"email"`
	AuthProvider AuthProvider `json:"authProvider"`
}

type CreateUserResponseDTO struct {
	Id           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"createdAt"`
	StreamAPIKey uuid.UUID `json:"streamAPIKey"`
	DisplayName  *string   `json:"displayName,omitempty"`
	PhoneNumber  *string   `json:"phoneNumber,omitempty"`
}

type AuthProvider string

const (
	ProviderGoogle AuthProvider = "google"
	ProviderLocal  AuthProvider = "local"
)
