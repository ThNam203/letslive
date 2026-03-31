package vod

import (
	"context"
	"sen1or/letslive/vod/domains"
	response "sen1or/letslive/vod/response"

	"github.com/gofrs/uuid/v5"
)

func (s *VODService) GetVODById(ctx context.Context, vodId uuid.UUID) (*domains.VOD, *response.Response[any]) {
	return s.vodRepo.GetById(ctx, vodId)
}
