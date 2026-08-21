package livestream

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
)

func (s *LivestreamService) GetRecommendedLivestreams(ctx context.Context, page int, limit int) ([]domains.Livestream, *response.Response[any]) {
	if page < 0 {
		page = 0
	}

	if limit <= 0 {
		limit = 10
	}

	if limit > 50 {
		limit = 50
	}

	livestreams, err := s.livestreamRepo.GetRecommendedLivestreams(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	if len(livestreams) == 0 {
		return livestreams, nil
	}

	userIdSet := make(map[uuid.UUID]struct{}, len(livestreams))
	for _, l := range livestreams {
		userIdSet[l.UserId] = struct{}{}
	}
	userIds := make([]uuid.UUID, 0, len(userIdSet))
	for id := range userIdSet {
		userIds = append(userIds, id)
	}

	statuses, statusErr := s.userGateway.GetUsersStatuses(ctx, userIds)
	if statusErr != nil {
		logger.Errorf(ctx, "failed to fetch author statuses, returning unfiltered results: %v", statusErr)
		return livestreams, nil
	}

	filtered := make([]domains.Livestream, 0, len(livestreams))
	for _, l := range livestreams {
		if statuses[l.UserId.String()] != "disabled" {
			filtered = append(filtered, l)
		}
	}

	return filtered, nil
}
