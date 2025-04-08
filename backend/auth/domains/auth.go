package domains

import (
	"context"
	serviceresponse "sen1or/letslive/auth/responses"
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
	GetByID(ctx context.Context, authId uuid.UUID) (*Auth, *serviceresponse.ServiceErrorResponse)
	GetByUserID(ctx context.Context, userId uuid.UUID) (*Auth, *serviceresponse.ServiceErrorResponse)
	GetByEmail(ctx context.Context, email string) (*Auth, *serviceresponse.ServiceErrorResponse)

	Create(ctx context.Context, auth Auth) (*Auth, *serviceresponse.ServiceErrorResponse)
	UpdatePasswordHash(ctx context.Context, authId, newPasswordHash string) *serviceresponse.ServiceErrorResponse
}
