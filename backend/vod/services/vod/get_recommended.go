package vod

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"
	response "sen1or/letslive/vod/response"

	"github.com/gofrs/uuid/v5"
)

func (s *VODService) GetRecommendedVODs(ctx context.Context, page int, limit int) ([]domains.VOD, *response.Response[any]) {
	if page < 0 {
		page = 0
	}

	if limit <= 0 {
		limit = 10
	}

	if limit > 50 {
		limit = 50
	}

	vods, err := s.vodRepo.GetPopular(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	if len(vods) == 0 {
		return vods, nil
	}

	userIdSet := make(map[uuid.UUID]struct{}, len(vods))
	for _, v := range vods {
		userIdSet[v.UserId] = struct{}{}
	}
	userIds := make([]uuid.UUID, 0, len(userIdSet))
	for id := range userIdSet {
		userIds = append(userIds, id)
	}

	statuses, statusErr := s.userGateway.GetUsersStatuses(ctx, userIds)
	if statusErr != nil {
		logger.Errorf(ctx, "failed to fetch author statuses, returning unfiltered results: %v", statusErr)
		return vods, nil
	}

	filtered := make([]domains.VOD, 0, len(vods))
	for _, v := range vods {
		if statuses[v.UserId.String()] != "disabled" {
			filtered = append(filtered, v)
		}
	}

	return filtered, nil
}
