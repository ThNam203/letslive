package domains

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	IsVerified   bool      `json:"isVerified" db:"is_verified"`
	IsOnline     bool      `json:"isOnline" db:"is_online"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	StreamAPIKey uuid.UUID `json:"streamAPIKey" db:"stream_api_key"` //TODO: hide stream api
}