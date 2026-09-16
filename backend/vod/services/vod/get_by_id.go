package vod

import (
	"context"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

func (s *VODService) GetVODById(ctx context.Context, vodId uuid.UUID) (*domains.VOD, error) {
	return s.vodRepo.GetById(ctx, vodId)
}
