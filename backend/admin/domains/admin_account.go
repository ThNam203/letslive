package domains

import (
	"context"
	"time"

	"sen1or/letslive/admin/response"

	"github.com/gofrs/uuid/v5"
)

type AdminAccount struct {
	Id           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}

type AdminAccountRepository interface {
	GetByEmail(ctx context.Context, email string) (*AdminAccount, *response.Response[any])
	GetByID(ctx context.Context, id uuid.UUID) (*AdminAccount, *response.Response[any])
}
