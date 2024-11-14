package domains

import (
	"github.com/gofrs/uuid/v5"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	IsVerified   bool      `json:"isVerified" db:"is_verified"`

	RefreshTokens []RefreshToken `json:"-"`
	VerifyTokens  []VerifyToken  `json:"-"`
}
