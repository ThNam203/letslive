package domains

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type RefreshToken struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Value     string     `json:"value" db:"value"`
	ExpiresAt time.Time  `json:"expiresAt" db:"expires_at"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	RevokedAt *time.Time `json:"revokedAt" db:"revoked_at"`
	UserID    uuid.UUID  `json:"userID" db:"user_id"`
}
