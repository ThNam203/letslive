package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

type VerifyToken struct {
	gorm.Model
	Token     string    `gorm:"type:varchar(255);unique;not null"`
	ExpiresAt time.Time `gorm:"not null"`

	UserID uuid.UUID `gorm:"not null;index"`
	User   User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type VerifyTokenRepository interface {
	CreateToken(userId uuid.UUID) (*VerifyToken, error)
	GetByToken(token string) (*VerifyToken, error)
	UpdateToken(VerifyToken) error
	DeleteToken(tokenId uint) error
}
