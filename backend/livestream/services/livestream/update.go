package livestream

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/dto"
	"sen1or/letslive/livestream/utils"

	"github.com/gofrs/uuid/v5"
)

func (s *LivestreamService) Update(ctx context.Context, data dto.UpdateLivestreamRequestDTO, streamId uuid.UUID, authorId uuid.UUID) (*domains.Livestream, error) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, domains.ErrInvalidPayload
	}

	currentLivestream, err := s.livestreamRepo.GetById(ctx, streamId)
	if err != nil {
		return nil, err
	}

	if currentLivestream.EndedAt != nil {
		return nil, domains.ErrLivestreamUpdateAfterEnded
	}

	if authorId != currentLivestream.UserId {
		return nil, domains.ErrForbidden
	}

	updated := false
	if data.Title != nil && *data.Title != currentLivestream.Title {
		currentLivestream.Title = *data.Title
		updated = true
	}
	if data.Description != nil && (currentLivestream.Description == nil || *data.Description != *currentLivestream.Description) {
		desc := *data.Description
		currentLivestream.Description = &desc
		updated = true
	}
	if data.ThumbnailURL != nil && (currentLivestream.ThumbnailURL == nil || *data.ThumbnailURL != *currentLivestream.ThumbnailURL) {
		thumb := *data.ThumbnailURL
		currentLivestream.ThumbnailURL = &thumb
		updated = true
	}
	if data.Visibility != nil && domains.LivestreamVisibility(*data.Visibility) != currentLivestream.Visibility {
		currentLivestream.Visibility = domains.LivestreamVisibility(*data.Visibility)
		updated = true
	}

	// only call update if changes were actually made
	if !updated {
		return currentLivestream, nil
	}

	updatedLivestream, err := s.livestreamRepo.Update(ctx, *currentLivestream)
	if err != nil {
		return nil, err
	}

	return updatedLivestream, nil
}
