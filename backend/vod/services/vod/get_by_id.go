package vod

import (
	"context"
	"sen1or/letslive/vod/domains"
	"sen1or/letslive/vod/pkg/logger"
	response "sen1or/letslive/vod/response"

	"github.com/gofrs/uuid/v5"
)

func (s *VODService) GetVODById(ctx context.Context, vodId uuid.UUID) (*domains.VOD, *response.Response[any]) {
	// TODO: this should not be the logic to increase view count, use it on video player actions
	err := s.vodRepo.IncrementViewCount(ctx, vodId)
	if err != nil {
		logger.Warnf(ctx, "failed to increase view count [getvodbyid id: %s]", vodId)
	}

	return s.vodRepo.GetById(ctx, vodId)
}
