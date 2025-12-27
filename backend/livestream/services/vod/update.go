package vod

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/dto"
	"sen1or/letslive/livestream/response"
	"sen1or/letslive/livestream/utils"

	"github.com/gofrs/uuid/v5"
)

func (s *VODService) UpdateVODMetadata(ctx context.Context, data dto.UpdateVODRequestDTO, vodId uuid.UUID, authorId uuid.UUID) (*domains.VOD, *response.Response[any]) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil)
	}

	currentVOD, err := s.vodRepo.GetById(ctx, vodId)
	if err != nil {
		return nil, err
	}

	if authorId != currentVOD.UserId {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_FORBIDDEN,
			nil,
			nil,
			nil,
		)
	}

	// TODO: Mapper: UpdateVODRequestDTOToVOD(data, currentVOD) -> domains.VOD
	updated := false
	if data.Title != nil && (*data.Title != currentVOD.Title) {
		currentVOD.Title = *data.Title
		updated = true
	}
	if data.Description != nil && (currentVOD.Description == nil || *data.Description != *currentVOD.Description) {
		currentVOD.Description = data.Description
		updated = true
	}
	if data.ThumbnailURL != nil && (currentVOD.ThumbnailURL == nil || *data.ThumbnailURL != *currentVOD.ThumbnailURL) {
		currentVOD.ThumbnailURL = data.ThumbnailURL
		updated = true
	}
	if data.Visibility != nil && domains.VODVisibility(*data.Visibility) != currentVOD.Visibility {
		currentVOD.Visibility = domains.VODVisibility(*data.Visibility)
		updated = true
	}

	if !updated {
		return currentVOD, nil
	}

	return s.vodRepo.Update(ctx, *currentVOD)
}
