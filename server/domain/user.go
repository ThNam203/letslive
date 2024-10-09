package domain

import (
	"sen1or/lets-live/server/util"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
)

type User struct {
	ID           uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Username     string    `json:"username" gorm:"unique;size:20;not null"`
	Email        string    `json:"email" gorm:"unique;not null"`
	PasswordHash string    `json:"-"`
	IsVerified   bool      `json:"isVerified" gorm:"not null;default:false"`
	IsOnline     bool      `json:"isOnline" gorm:"not null;default:false"`
	CreatedAt    time.Time `json:"createdAt" gorm:"default:current_timestamp"`
	StreamAPIKey uuid.UUID `json:"streamAPIKey" gorm:"type:uuid;not null;default:uuid_generate_v4()"`

	RefreshTokens []RefreshToken `json:"-"`
	VerifyTokens  []VerifyToken  `json:"-"`
}

func (u *User) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(&u)

	if err != nil {
		util.LogValidationErrors(err)
		return err
	}

	return nil
}

type UserRepository interface {
	GetByID(uuid.UUID) (*User, error)
	GetByName(string) (*User, error)
	GetByEmail(string) (*User, error)
	GetByAPIKey(uuid.UUID) (*User, error)
	GetStreamingUsers() ([]User, error)

	Create(User) error
	Update(User) error
	Save(User) error
	Delete(string) error
}
