package dto

import (
	"testing"

	"sen1or/letslive/livestream/utils"
)

func TestEndLivestreamRequestDTOVODStatus(t *testing.T) {
	tests := []struct {
		status string
		valid  bool
	}{
		{status: "", valid: true},
		{status: "ready", valid: true},
		{status: "processing", valid: true},
		{status: "failed", valid: false},
		{status: "live", valid: false},
	}
	for _, tt := range tests {
		err := utils.Validator.Struct(EndLivestreamRequestDTO{VODStatus: tt.status})
		if (err == nil) != tt.valid {
			t.Errorf("vodStatus %q: err = %v, want valid %v", tt.status, err, tt.valid)
		}
	}
}
