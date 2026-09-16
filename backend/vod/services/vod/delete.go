package vod

import (
	"context"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

func (s *VODService) Delete(ctx context.Context, vodId uuid.UUID, authorId uuid.UUID) error {
	vod, err := s.vodRepo.GetById(ctx, vodId)
	if err != nil {
		return err
	}

	if authorId != vod.UserId {
		return domains.ErrForbidden
	}

	return s.vodRepo.Delete(ctx, vodId)
}
