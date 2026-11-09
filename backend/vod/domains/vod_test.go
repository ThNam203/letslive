package domains

import (
	"testing"

	"github.com/gofrs/uuid/v5"
)

// Only a ready public VOD is shown to people other than its owner; the owner
// sees every one of theirs, whatever its visibility or status.
func TestVODVisibleTo(t *testing.T) {
	owner := uuid.Must(uuid.NewV4())
	stranger := uuid.Must(uuid.NewV4())

	tests := []struct {
		visibility VODVisibility
		status     VODStatus
		viewer     *uuid.UUID
		want       bool
	}{
		{VODPublicVisibility, VODStatusReady, nil, true},
		{VODPublicVisibility, VODStatusReady, &stranger, true},
		{VODPublicVisibility, VODStatusProcessing, nil, false},
		{VODPublicVisibility, VODStatusProcessing, &stranger, false},
		{VODPublicVisibility, VODStatusUploading, &stranger, false},
		{VODPublicVisibility, VODStatusFailed, &stranger, false},
		{VODPrivateVisibility, VODStatusReady, nil, false},
		{VODPrivateVisibility, VODStatusReady, &stranger, false},
		{VODPublicVisibility, VODStatusProcessing, &owner, true},
		{VODPublicVisibility, VODStatusFailed, &owner, true},
		{VODPrivateVisibility, VODStatusProcessing, &owner, true},
		{VODPrivateVisibility, VODStatusReady, &owner, true},
	}

	for _, tt := range tests {
		vod := &VOD{UserId: owner, Visibility: tt.visibility, Status: tt.status}
		who := "signed out"
		switch tt.viewer {
		case &owner:
			who = "owner"
		case &stranger:
			who = "stranger"
		}
		if got := vod.VisibleTo(tt.viewer); got != tt.want {
			t.Errorf("%s %s vod, %s: VisibleTo = %v, want %v", tt.visibility, tt.status, who, got, tt.want)
		}
	}
}
