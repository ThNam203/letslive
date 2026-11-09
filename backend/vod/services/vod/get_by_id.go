package vod

import (
	"context"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

// GetVODById returns a VOD the caller may see (domains.VOD.VisibleTo).
// viewerId is the verified caller, or nil when signed out. Anyone else gets
// ErrVODNotFound so a private or unready VOD can't be told apart from a
// missing one.
func (s *VODService) GetVODById(ctx context.Context, vodId uuid.UUID, viewerId *uuid.UUID) (*domains.VOD, error) {
	vod, err := s.vodRepo.GetById(ctx, vodId)
	if err != nil {
		return nil, err
	}
	if !vod.VisibleTo(viewerId) {
		return nil, domains.ErrVODNotFound
	}
	return vod, nil
}
