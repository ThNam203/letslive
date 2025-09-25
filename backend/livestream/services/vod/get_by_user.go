package vod

import (
	"context"
	"sen1or/letslive/livestream/domains"
	response "sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (s *VODService) GetAllVODsByUser(ctx context.Context, userId uuid.UUID, page int, limit int) ([]domains.VOD, *response.Response[any]) {
	if page < 0 {
		page = 0
	}

	if limit <= 0 {
		limit = 10
	}

	return s.vodRepo.GetByUser(ctx, userId, page, limit)
}
