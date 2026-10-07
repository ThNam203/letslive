package utils

import (
	"unicode"
	"unicode/utf16"

	"github.com/go-playground/validator/v10"
)

// in UTF-16 units, as the web form counts
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

	length := len(utf16.Encode([]rune(password)))
	if length < passwordMinLength || length > passwordMaxLength {
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
