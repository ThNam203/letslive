package utils

import (
	"unicode"

	"github.com/go-playground/validator/v10"
)

const (
	passwordMinLength = 8
	passwordMaxLength = 72
)

var Validator = validator.New(validator.WithRequiredStructEnabled())

func init() {
	// register custom password validation, ignore error
	_ = Validator.RegisterValidation("password", validatePassword)
}

// - at least 8 characters
// - at least one lowercase letter
// - at least one uppercase letter
// - at least one special character
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// check minimum length
	if len(password) < passwordMinLength {
		return false
	}

	// check maximum length
	if len(password) > passwordMaxLength {
		return false
	}

	hasLower := false
	hasUpper := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case !unicode.IsLetter(char) && !unicode.IsNumber(char):
			hasSpecial = true
		}
	}

	return hasLower && hasUpper && hasSpecial
}
