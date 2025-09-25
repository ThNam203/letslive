package vod

import (
	"context"
	response "sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (s *VODService) IncrementVODView(ctx context.Context, vodId uuid.UUID) *response.Response[any] {
	// TODO: Add rate limiting
	return s.vodRepo.IncrementViewCount(ctx, vodId)
}
