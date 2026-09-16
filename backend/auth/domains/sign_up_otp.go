package domains

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

type SignUpOTP struct {
	Id        uuid.UUID  `json:"id" db:"id"`
	Code      string     `json:"code" db:"code"`
	Email     string     `json:"email" db:"email"`
	ExpiresAt time.Time  `json:"expiresAt" db:"expires_at"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	UsedAt    *time.Time `json:"usedAt" db:"used_at"`
}

type SignUpOTPRepository interface {
	Insert(ctx context.Context, newOTP SignUpOTP) error
	GetOTP(ctx context.Context, code, email string) (*SignUpOTP, error)
	UpdateUsedAt(ctx context.Context, otpId uuid.UUID, usedAt time.Time) error
}
