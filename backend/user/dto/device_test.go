package dto

import (
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
)

func strPtr(s string) *string { return &s }

func TestInvalidPlatformReturnsValidationError(t *testing.T) {
	validate := validator.New()
	cases := []struct {
		name    string
		req     RegisterDeviceRequestDTO
		wantErr bool
	}{
		{"ios ok", RegisterDeviceRequestDTO{Token: "abc", Platform: "ios"}, false},
		{"android with locale ok", RegisterDeviceRequestDTO{Token: "abc", Platform: "android", Locale: strPtr("vi-VN")}, false},
		{"invalid platform", RegisterDeviceRequestDTO{Token: "abc", Platform: "windows"}, true},
		{"empty platform", RegisterDeviceRequestDTO{Token: "abc"}, true},
		{"empty token", RegisterDeviceRequestDTO{Platform: "ios"}, true},
		{"token too long", RegisterDeviceRequestDTO{Token: strings.Repeat("a", MaxDeviceTokenLength+1), Platform: "ios"}, true},
		{"token at max length", RegisterDeviceRequestDTO{Token: strings.Repeat("a", MaxDeviceTokenLength), Platform: "ios"}, false},
		{"bad locale", RegisterDeviceRequestDTO{Token: "abc", Platform: "ios", Locale: strPtr("not a locale!")}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validate.Struct(tc.req)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
