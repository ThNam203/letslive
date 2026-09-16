package vod

import (
	"context"
	"sen1or/letslive/vod/domains"
)

func (s *VODService) Create(ctx context.Context, vod domains.VOD) (*domains.VOD, error) {
	return s.vodRepo.Create(ctx, vod)
}
