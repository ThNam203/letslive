package vod

import (
	"context"
	"sen1or/letslive/vod/domains"
	"sen1or/letslive/vod/dto"
	"sen1or/letslive/vod/utils"

	"github.com/gofrs/uuid/v5"
)

func (s *VODService) UpdateVODMetadata(ctx context.Context, data dto.UpdateVODRequestDTO, vodId uuid.UUID, authorId uuid.UUID) (*domains.VOD, error) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, domains.ErrInvalidPayload
	}

	currentVOD, err := s.vodRepo.GetById(ctx, vodId)
	if err != nil {
		return nil, err
	}

	if authorId != currentVOD.UserId {
		return nil, domains.ErrForbidden
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
