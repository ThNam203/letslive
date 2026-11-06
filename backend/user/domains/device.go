package domains

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// DeviceToken is a push-notification (FCM) registration token owned by one user.
type DeviceToken struct {
	Id         uuid.UUID `json:"id" db:"id"`
	UserId     uuid.UUID `json:"userId" db:"user_id"`
	Token      string    `json:"token" db:"token"`
	Platform   string    `json:"platform" db:"platform"`
	Locale     *string   `json:"locale" db:"locale"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	LastSeenAt time.Time `json:"lastSeenAt" db:"last_seen_at"`
}

type DeviceTokenRepository interface {
	// Upsert stores the token for token.UserId. When the token already exists
	// it is reassigned to token.UserId and platform, locale and last_seen_at
	// are refreshed.
	Upsert(ctx context.Context, token DeviceToken) (*DeviceToken, error)
	ListTokensByUserID(ctx context.Context, userID uuid.UUID) ([]DeviceToken, error)
	// DeleteByToken removes a token regardless of owner. Used by the push
	// sender for tokens FCM reports as unregistered.
	DeleteByToken(ctx context.Context, token string) error
	// DeleteByUserAndToken removes the token only when it belongs to userID.
	// Missing or foreign tokens are a no-op.
	DeleteByUserAndToken(ctx context.Context, userID uuid.UUID, token string) error
}
