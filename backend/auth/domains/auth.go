package domains

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

type Auth struct {
	Id           uuid.UUID  `json:"id" db:"id"`
	UserId       *uuid.UUID `json:"userID" db:"user_id"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
}

type AuthRepository interface {
	GetByID(ctx context.Context, authId uuid.UUID) (*Auth, error)
	GetByUserID(ctx context.Context, userId uuid.UUID) (*Auth, error)
	GetByEmail(ctx context.Context, email string) (*Auth, error)

	Create(ctx context.Context, auth Auth) (*Auth, error)
	UpdatePasswordHash(ctx context.Context, authId, newPasswordHash string) error
}
