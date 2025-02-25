package domains

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type User struct {
	Id                uuid.UUID `json:"id" db:"id"`
	Username          string    `json:"username" db:"username"`
	Email             string    `json:"email" db:"email"`
	IsVerified        bool      `json:"isVerified" db:"is_verified"`
	IsOnline          bool      `json:"isOnline" db:"is_online"`
	IsActive          bool      `json:"isActive" db:"is_active"`
	CreatedAt         time.Time `json:"createdAt" db:"created_at"`
	StreamAPIKey      uuid.UUID `json:"streamAPIKey" db:"stream_api_key"`
	DisplayName       *string   `json:"displayName,omitempty" db:"display_name"`
	PhoneNumber       *string   `json:"phoneNumber,omitempty" db:"phone_number"`
	Bio               *string   `json:"bio,omitempty" db:"bio"`
	ProfilePicture    *string   `json:"profilePicture,omitempty" db:"profile_picture"`
	BackgroundPicture *string   `json:"backgroundPicture,omitempty" db:"background_picture"`

	LivestreamInformation `json:"livestreamInformation"`
}
