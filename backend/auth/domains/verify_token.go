package domains

import (
	servererrors "sen1or/letslive/auth/errors"
	"time"

	"github.com/gofrs/uuid/v5"
)

type VerifyToken struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expiresAt" db:"expires_at"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
}

type VerifyTokenRepository interface {
	Insert(newToken VerifyToken) *servererrors.ServerError
	GetByValue(token string) (*VerifyToken, *servererrors.ServerError)
	DeleteByID(uuid.UUID) *servererrors.ServerError
	DeleteByValue(token string) *servererrors.ServerError
	DeleteAllOfUser(userId uuid.UUID) *servererrors.ServerError
}
