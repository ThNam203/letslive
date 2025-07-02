package vod

import (
	"context"
	serviceresponse "sen1or/letslive/livestream/responses"

	"github.com/gofrs/uuid/v5"
)

func (s *VODService) Delete(ctx context.Context, vodId uuid.UUID, authorId uuid.UUID) *serviceresponse.ServiceErrorResponse {
	vod, err := s.vodRepo.GetById(ctx, vodId)
	if err != nil {
		return err
	}

	if authorId != vod.UserId {
		return serviceresponse.ErrForbidden
	}

	return s.vodRepo.Delete(ctx, vodId)
}
