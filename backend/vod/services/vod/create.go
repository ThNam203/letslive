package vod

import (
	"context"
	"sen1or/letslive/vod/domains"
	"sen1or/letslive/vod/response"
)

func (s *VODService) Create(ctx context.Context, vod domains.VOD) (*domains.VOD, *response.Response[any]) {
	return s.vodRepo.Create(ctx, vod)
}
