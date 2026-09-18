package dto_test

import (
	"testing"

	"sen1or/letslive/user/dto"

	"github.com/go-playground/validator/v10"
)

// Notification action URLs are app-relative paths (e.g. "/users/<id>/gifts"),
// resolved by the web client. The "url" tag rejects those because they carry
// no scheme, so the DTO must validate them as URI references instead.
func TestCreateNotificationRequestDTO_ActionUrl(t *testing.T) {
	const userId = "0f8fad5b-d9cb-469f-a165-70867728950e"

	tests := []struct {
		name      string
		actionUrl string
		wantValid bool
	}{
		{"relative path used by gift notifications", "/users/" + userId + "/gifts", true},
		{"relative path with query", "/users/" + userId + "/gifts?page=2", true},
		{"absolute url", "https://letslive.test/users/me/gifts", true},
		{"not a uri reference", "users/me/gifts", false},
		{"empty string", "", false},
	}

	validate := validator.New()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actionUrl := tc.actionUrl
			req := dto.CreateNotificationRequestDTO{
				UserId:    userId,
				Type:      "gift_received",
				Title:     "You received a gift!",
				Message:   "someone sent you a gift",
				ActionUrl: &actionUrl,
			}

			err := validate.Struct(req)
			if tc.wantValid && err != nil {
				t.Fatalf("expected actionUrl %q to be valid, got %v", tc.actionUrl, err)
			}
			if !tc.wantValid && err == nil {
				t.Fatalf("expected actionUrl %q to be rejected", tc.actionUrl)
			}
		})
	}
}
