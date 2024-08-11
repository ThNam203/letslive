package domain

import (
	"sen1or/lets-live/server/util"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID           uuid.UUID `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"unique;size:20;not null"`
	Email        string    `json:"email" gorm:"unique;not null"`
	PasswordHash string    `gorm:"not null" validate:"required"`
}

func (u *User) Validate() error {
	// TODO: make validate variable singleton
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(&u)

	if err != nil {
		util.LogValidationErrors(err)
		return err
	}

	return nil
}

type UserRepository interface {
	GetByID(string) (*User, error)
	GetByName(string) (*User, error)
	GetByEmail(string) (*User, error)

	Create(User) error
	Update(User) error
	Delete(string) error
}
