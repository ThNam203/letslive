package domain

import (
	"sen1or/lets-live/server/util"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID           uuid.UUID `gorm:"primaryKey" validate:"required,uuid"`
	Username     string    `gorm:"unique;size:20;not null" validate:"required,gte=6,lte=20"`
	Email        string    `gorm:"unique;not null" validate:"required,email"`
	PasswordHash string    `gorm:"not null" validate:"required"`
}

var validate *validator.Validate = validator.New(validator.WithRequiredStructEnabled())

func (u *User) Validate() error {
	err := validate.Struct(&u)

	if err != nil {
		util.LogValidationErrors(err)
		return err
	}

	return nil
}

type UserRepository interface {
	GetByID(string) (User, error)
	GetByName(string) (User, error)
	Create(*User) error
	Update(*User) error
	Delete(string) error
}
