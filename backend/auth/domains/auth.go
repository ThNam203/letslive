package domains

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type Auth struct {
	Id           uuid.UUID `json:"id" db:"id"`
	UserId       uuid.UUID `json:"userID" db:"user_id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}
