package domains

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// TODO: check if ID has any of use, if not just use UserId as primary key
type Auth struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"userID" db:"user_id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	IsVerified   bool      `json:"isVerified" db:"is_verified"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}
