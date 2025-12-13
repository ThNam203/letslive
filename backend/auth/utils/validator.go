package utils

import (
	"unicode"

	"sen1or/letslive/auth/constants"

	"github.com/go-playground/validator/v10"
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
	if len(password) < constants.PASSWORD_MIN_LENGTH {
		return false
	}

	// check maximum length
	if len(password) > constants.PASSWORD_MAX_LENGTH {
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
