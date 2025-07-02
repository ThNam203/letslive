package vod

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/pkg/logger"
	serviceresponse "sen1or/letslive/livestream/responses"

	"github.com/gofrs/uuid/v5"
)

func (s *VODService) GetVODById(ctx context.Context, vodId uuid.UUID) (*domains.VOD, *serviceresponse.ServiceErrorResponse) {
	// TODO: this should not be the logic to increase view count, use it on video player actions
	err := s.vodRepo.IncrementViewCount(ctx, vodId)
	if err != nil {
		logger.Warnf("failed to increase view count [getvodbyid id: %s]", vodId)
	}

	return s.vodRepo.GetById(ctx, vodId)
}
