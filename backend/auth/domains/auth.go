package domains

import (
	servererrors "sen1or/letslive/auth/errors"
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

type AuthRepository interface {
	GetByID(uuid.UUID) (*Auth, *servererrors.ServerError)
	GetByUserID(uuid.UUID) (*Auth, *servererrors.ServerError)
	GetByEmail(string) (*Auth, *servererrors.ServerError)

	Create(Auth) (*Auth, *servererrors.ServerError)
	UpdatePasswordHash(authId, newPasswordHash string) *servererrors.ServerError
	Delete(uuid.UUID) *servererrors.ServerError
}
