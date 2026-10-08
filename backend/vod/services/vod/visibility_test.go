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

func newService(visibility domains.VODVisibility, ownerId uuid.UUID) (*VODService, *vodRepoStub) {
	repo := &vodRepoStub{vod: &domains.VOD{Id: uuid.Must(uuid.NewV4()), UserId: ownerId, Visibility: visibility, Duration: 600}}
	return &VODService{vodRepo: repo}, repo
}

func TestGetVODById_visibility(t *testing.T) {
	owner := uuid.Must(uuid.NewV4())
	stranger := uuid.Must(uuid.NewV4())

	tests := []struct {
		name       string
		visibility domains.VODVisibility
		viewer     *uuid.UUID
		wantErr    error
	}{
		{"public vod is visible when signed out", domains.VODPublicVisibility, nil, nil},
		{"public vod is visible to a stranger", domains.VODPublicVisibility, &stranger, nil},
		{"private vod is hidden when signed out", domains.VODPrivateVisibility, nil, domains.ErrVODNotFound},
		{"private vod is hidden from a stranger", domains.VODPrivateVisibility, &stranger, domains.ErrVODNotFound},
		{"private vod is visible to its owner", domains.VODPrivateVisibility, &owner, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := newService(tt.visibility, owner)

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
		viewer     *uuid.UUID
		wantErr    error
		wantViews  int
	}{
		{"public vod counts a signed-out view", domains.VODPublicVisibility, nil, nil, 1},
		{"private vod ignores a signed-out view", domains.VODPrivateVisibility, nil, domains.ErrVODNotFound, 0},
		{"private vod ignores a stranger's view", domains.VODPrivateVisibility, &stranger, domains.ErrVODNotFound, 0},
		{"private vod counts its owner's view", domains.VODPrivateVisibility, &owner, nil, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := newService(tt.visibility, owner)

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
