package vod

import (
	"context"
	serviceresponse "sen1or/letslive/livestream/responses"

	"github.com/gofrs/uuid/v5"
)

func (s *VODService) IncrementVODView(ctx context.Context, vodId uuid.UUID) *serviceresponse.ServiceErrorResponse {
	// TODO: Add rate limiting
	return s.vodRepo.IncrementViewCount(ctx, vodId)
}
