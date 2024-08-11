package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Value     string    `gorm:"type:varchar(255);unique;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:current_timestamp"`
	Revoked   bool      `gorm:"default:false;not null"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type RefreshTokenRepository interface {
	RevokeByValue(string) error
	Create(RefreshToken) error
	FindByValue(string) (*RefreshToken, error)
	GenerateTokenPair(userId uuid.UUID) (refreshToken string, accessToken string, err error)
}
