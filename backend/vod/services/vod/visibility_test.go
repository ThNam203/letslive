package vod

import (
	"context"
	"errors"
	"testing"

	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

// vodRepoStub serves one VOD and counts view increments; every other method
// panics through the nil embedded interface, which these tests never reach.
type vodRepoStub struct {
	domains.VODRepository
	vod   *domains.VOD
	views int
}

func (r *vodRepoStub) GetById(context.Context, uuid.UUID) (*domains.VOD, error) {
	return r.vod, nil
}

func (r *vodRepoStub) IncrementViewCount(context.Context, uuid.UUID) error {
	r.views++
	return nil
}

func newService(visibility domains.VODVisibility, status domains.VODStatus, ownerId uuid.UUID) (*VODService, *vodRepoStub) {
	repo := &vodRepoStub{vod: &domains.VOD{Id: uuid.Must(uuid.NewV4()), UserId: ownerId, Visibility: visibility, Status: status, Duration: 600}}
	return &VODService{vodRepo: repo}, repo
}

func TestGetVODById_visibility(t *testing.T) {
	owner := uuid.Must(uuid.NewV4())
	stranger := uuid.Must(uuid.NewV4())

	tests := []struct {
		name       string
		visibility domains.VODVisibility
		status     domains.VODStatus
		viewer     *uuid.UUID
		wantErr    error
	}{
		{"public vod is visible when signed out", domains.VODPublicVisibility, domains.VODStatusReady, nil, nil},
		{"public vod is visible to a stranger", domains.VODPublicVisibility, domains.VODStatusReady, &stranger, nil},
		{"private vod is hidden when signed out", domains.VODPrivateVisibility, domains.VODStatusReady, nil, domains.ErrVODNotFound},
		{"private vod is hidden from a stranger", domains.VODPrivateVisibility, domains.VODStatusReady, &stranger, domains.ErrVODNotFound},
		{"private vod is visible to its owner", domains.VODPrivateVisibility, domains.VODStatusReady, &owner, nil},
		{"processing public vod is hidden from a stranger", domains.VODPublicVisibility, domains.VODStatusProcessing, &stranger, domains.ErrVODNotFound},
		{"failed public vod is hidden when signed out", domains.VODPublicVisibility, domains.VODStatusFailed, nil, domains.ErrVODNotFound},
		{"processing public vod is visible to its owner", domains.VODPublicVisibility, domains.VODStatusProcessing, &owner, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := newService(tt.visibility, tt.status, owner)

			got, err := svc.GetVODById(context.Background(), repo.vod.Id, tt.viewer)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && got == nil {
				t.Fatal("expected the vod")
			}
			if tt.wantErr != nil && got != nil {
				t.Fatal("a hidden vod must not be returned")
			}
		})
	}
}

func TestRegisterView_visibility(t *testing.T) {
	owner := uuid.Must(uuid.NewV4())
	stranger := uuid.Must(uuid.NewV4())

	tests := []struct {
		name       string
		visibility domains.VODVisibility
		status     domains.VODStatus
		viewer     *uuid.UUID
		wantErr    error
		wantViews  int
	}{
		{"public vod counts a signed-out view", domains.VODPublicVisibility, domains.VODStatusReady, nil, nil, 1},
		{"processing vod ignores a stranger's view", domains.VODPublicVisibility, domains.VODStatusProcessing, &stranger, domains.ErrVODNotFound, 0},
		{"private vod ignores a signed-out view", domains.VODPrivateVisibility, domains.VODStatusReady, nil, domains.ErrVODNotFound, 0},
		{"private vod ignores a stranger's view", domains.VODPrivateVisibility, domains.VODStatusReady, &stranger, domains.ErrVODNotFound, 0},
		{"private vod counts its owner's view", domains.VODPrivateVisibility, domains.VODStatusReady, &owner, nil, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := newService(tt.visibility, tt.status, owner)

			err := svc.RegisterView(context.Background(), repo.vod.Id, tt.viewer, 600)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if repo.views != tt.wantViews {
				t.Fatalf("views = %d, want %d", repo.views, tt.wantViews)
			}
		})
	}
}
