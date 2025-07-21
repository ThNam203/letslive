package livestream

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/responses"
)

func (s *LivestreamService) GetRecommendedLivestreams(ctx context.Context, page int, limit int) ([]domains.Livestream, *serviceresponse.ServiceErrorResponse) {
	if page < 0 {
		page = 0
	}

	if limit <= 0 {
		limit = 10
	}

	if limit > 50 {
		limit = 50
	}

	return s.livestreamRepo.GetRecommendedLivestreams(ctx, page, limit)
}
