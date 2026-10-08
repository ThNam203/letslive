package dto

import (
	"strings"
	"testing"

	"sen1or/letslive/auth/utils"
)

func TestSignUpUsernameLength(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{"5 characters", strings.Repeat("a", 5), false},
		{"6 characters", strings.Repeat("a", 6), true},
		{"30 characters", strings.Repeat("a", 30), true},
		{"31 characters", strings.Repeat("a", 31), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := SignUpRequestDTO{Username: tt.username, Email: "a@b.co", Password: "Password123!", OTPCode: "123456"}
			err := utils.Validator.Struct(&form)
			if (err == nil) != tt.want {
				t.Fatalf("valid = %v, want %v (err: %v)", err == nil, tt.want, err)
			}
		})
	}
}

// login and the current password are only compared with the stored hash, so
// a password made under older rules must still get through
func TestStoredPasswordsAreOnlyLengthChecked(t *testing.T) {
	older := "Ữữ@ữ" // 4 characters, accepted when the minimum counted bytes

	if err := utils.Validator.Struct(&LogInRequestDTO{Email: "a@b.co", Password: older}); err != nil {
		t.Fatalf("login rejected an older password: %v", err)
	}
	if err := utils.Validator.Struct(&ChangePasswordRequestDTO{OldPassword: older, NewPassword: "Password123!"}); err != nil {
		t.Fatalf("change password rejected an older current password: %v", err)
	}
	if err := utils.Validator.Struct(&LogInRequestDTO{Email: "a@b.co", Password: strings.Repeat("a", 73)}); err == nil {
		t.Fatal("login accepted a 73 character password")
	}
}
