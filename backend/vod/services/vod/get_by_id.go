package vod

import (
	"context"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

// GetVODById returns a public VOD, or a private one to its owner. viewerId is
// the verified caller, or nil when signed out. Anyone else gets ErrVODNotFound
// so a private VOD can't be told apart from a missing one.
func (s *VODService) GetVODById(ctx context.Context, vodId uuid.UUID, viewerId *uuid.UUID) (*domains.VOD, error) {
	vod, err := s.vodRepo.GetById(ctx, vodId)
	if err != nil {
		return nil, err
	}
	if !canView(vod, viewerId) {
		return nil, domains.ErrVODNotFound
	}
	return vod, nil
}

func canView(vod *domains.VOD, viewerId *uuid.UUID) bool {
	if vod.Visibility == domains.VODPublicVisibility {
		return true
	}
	return viewerId != nil && *viewerId == vod.UserId
}
