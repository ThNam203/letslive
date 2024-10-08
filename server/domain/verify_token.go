package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type VerifyToken struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Token     string    `gorm:"type:varchar(255);unique;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"default:current_timestamp"`

	UserID uuid.UUID `gorm:"not null;index"`
	User   User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type VerifyTokenRepository interface {
	CreateToken(userId uuid.UUID) (*VerifyToken, error)
	GetByToken(token string) (*VerifyToken, error)
	UpdateToken(VerifyToken) error
	DeleteToken(uuid.UUID) error
}
