package utils

import (
	"strings"
	"testing"
)

func TestPasswordRule(t *testing.T) {
	type form struct {
		Password string `validate:"password"`
	}

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"8 characters", "Passw0r!", true},
		{"7 characters", "Pass0r!", false},
		{"72 characters", "Aa!" + strings.Repeat("x", 69), true},
		{"73 characters", "Aa!" + strings.Repeat("x", 70), false},
		{"72 vietnamese characters", "Ữữ!" + strings.Repeat("ữ", 69), true},
		{"73 vietnamese characters", "Ữữ!" + strings.Repeat("ữ", 70), false},
		{"4 characters but 10 bytes", "Ữữ@ữ", false},
		{"6 characters but 9 UTF-16 units", "Aa!😀😀😀", true},
		{"72 UTF-16 units with emoji", "Aa!" + strings.Repeat("😀", 34) + "x", true},
		{"73 UTF-16 units with emoji", "Aa!" + strings.Repeat("😀", 35), false},
		{"no uppercase", "password1!", false},
		{"no lowercase", "PASSWORD1!", false},
		{"no special character", "Password1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validator.Struct(form{Password: tt.password})
			if (err == nil) != tt.want {
				t.Fatalf("valid = %v, want %v (err: %v)", err == nil, tt.want, err)
			}
		})
	}
}
