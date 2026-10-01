package user

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"sen1or/letslive/user/services"
)

func newCropRequest(t *testing.T, fields map[string]string) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for name, value := range fields {
		if err := writer.WriteField(name, value); err != nil {
			t.Fatalf("write field: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, "/v1/user/me/profile-picture", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestParseAvatarCrop(t *testing.T) {
	cases := []struct {
		name    string
		fields  map[string]string
		want    *services.AvatarCrop
		wantErr bool
	}{
		{"no crop", map[string]string{}, nil, false},
		{"full crop", map[string]string{"crop-x": "10", "crop-y": "20", "crop-size": "300"}, &services.AvatarCrop{X: 10, Y: 20, Size: 300}, false},
		{"missing size", map[string]string{"crop-x": "10", "crop-y": "20"}, nil, true},
		{"only size", map[string]string{"crop-size": "300"}, nil, true},
		{"not a number", map[string]string{"crop-x": "ten", "crop-y": "20", "crop-size": "300"}, nil, true},
		{"fractional", map[string]string{"crop-x": "10.5", "crop-y": "20", "crop-size": "300"}, nil, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseAvatarCrop(newCropRequest(t, tc.fields))
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("crop = %+v, want %+v", got, tc.want)
			}
		})
	}
}
