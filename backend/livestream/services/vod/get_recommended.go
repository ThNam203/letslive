package vod

import (
	"context"
	"sen1or/letslive/livestream/domains"
	serviceresponse "sen1or/letslive/livestream/responses"
)

func (s *VODService) GetRecommendedVODs(ctx context.Context, page int, limit int) ([]domains.VOD, *serviceresponse.ServiceErrorResponse) {
	if page < 0 {
		page = 0
	}

	if limit <= 0 {
		limit = 10
	}

	if limit > 50 {
		limit = 50
	}

	return s.vodRepo.GetPopular(ctx, page, limit)
}
