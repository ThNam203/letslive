package domains

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	IsVerified   bool      `json:"isVerified" db:"is_verified"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`

	RefreshTokens []RefreshToken `json:"-"`
	VerifyTokens  []VerifyToken  `json:"-"`
}
