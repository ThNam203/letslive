package vod

import (
	"context"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (s *VODService) Delete(ctx context.Context, vodId uuid.UUID, authorId uuid.UUID) *response.Response[any] {
	vod, err := s.vodRepo.GetById(ctx, vodId)
	if err != nil {
		return err
	}

	if authorId != vod.UserId {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_FORBIDDEN,
			nil,
			nil,
			nil,
		)
	}

	return s.vodRepo.Delete(ctx, vodId)
}
