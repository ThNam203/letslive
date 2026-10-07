package dto

import (
	"strings"
	"testing"

	"sen1or/letslive/vod/utils"
)

func TestUpdateVODRequestDTOValidation(t *testing.T) {
	ptr := func(s string) *string { return &s }

	tests := []struct {
		name string
		dto  UpdateVODRequestDTO
		ok   bool
	}{
		{"nothing to update", UpdateVODRequestDTO{}, true},
		{"one character title", UpdateVODRequestDTO{Title: ptr("x")}, true},
		{"255 character title", UpdateVODRequestDTO{Title: ptr(strings.Repeat("a", 255))}, true},
		{"255 multi-byte character title", UpdateVODRequestDTO{Title: ptr(strings.Repeat("ữ", 255))}, true},
		{"empty title", UpdateVODRequestDTO{Title: ptr("")}, false},
		{"256 character title", UpdateVODRequestDTO{Title: ptr(strings.Repeat("a", 256))}, false},
		{"empty description", UpdateVODRequestDTO{Description: ptr("")}, true},
		{"1000 character description", UpdateVODRequestDTO{Description: ptr(strings.Repeat("a", 1000))}, true},
		{"1001 character description", UpdateVODRequestDTO{Description: ptr(strings.Repeat("a", 1001))}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.Validator.Struct(&tt.dto)
			if (err == nil) != tt.ok {
				t.Fatalf("valid = %v, want %v (err: %v)", err == nil, tt.ok, err)
			}
		})
	}
}
