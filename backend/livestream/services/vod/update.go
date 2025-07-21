package vod

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/dto"
	serviceresponse "sen1or/letslive/livestream/responses"
	"sen1or/letslive/livestream/utils"

	"github.com/gofrs/uuid/v5"
)

func (s *VODService) UpdateVODMetadata(ctx context.Context, data dto.UpdateVODRequestDTO, vodId uuid.UUID, authorId uuid.UUID) (*domains.VOD, *serviceresponse.ServiceErrorResponse) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, serviceresponse.ErrInvalidPayload
	}

	currentVOD, err := s.vodRepo.GetById(ctx, vodId)
	if err != nil {
		return nil, err
	}

	if authorId != currentVOD.UserId {
		return nil, serviceresponse.ErrForbidden
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
	if data.Visibility != nil && domains.VODVisibility(*data.Visibility) != currentVOD.Visibility {
		currentVOD.Visibility = domains.VODVisibility(*data.Visibility)
		updated = true
	}

	if !updated {
		return currentVOD, nil
	}

	return s.vodRepo.Update(ctx, *currentVOD)
}
