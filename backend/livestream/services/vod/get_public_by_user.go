package vod

import (
	"context"
	"sen1or/letslive/livestream/domains"
	serviceresponse "sen1or/letslive/livestream/responses"

	"github.com/gofrs/uuid/v5"
)

func (s *VODService) GetPublicVODsByUser(ctx context.Context, userId uuid.UUID, page int, limit int) ([]domains.VOD, *serviceresponse.ServiceErrorResponse) {
	if page < 0 {
		page = 0
	}

	if limit <= 0 {
		limit = 10
	}

	return s.vodRepo.GetPublicVODsByUser(ctx, userId, page, limit)
}
