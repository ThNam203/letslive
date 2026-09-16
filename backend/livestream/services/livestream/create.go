package livestream

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/dto"
	"sen1or/letslive/livestream/utils"
	"sen1or/letslive/shared/pkg/logger"
	"time"
)

func (s *LivestreamService) Create(ctx context.Context, data dto.CreateLivestreamRequestDTO) (*domains.Livestream, error) {
	if err := utils.Validator.Struct(&data); err != nil {
		logger.Debugf(ctx, "create livestream validation failed: %s", err.Error())
		return nil, domains.ErrInvalidPayload
	}

	var titleString = ""
	if data.Title != nil {
		titleString = *data.Title
	}

	if data.Visibility == nil {
		*data.Visibility = domains.LivestreamPublicVisibility
	}

	livestreamData := domains.Livestream{
		UserId:       data.UserId,
		Title:        titleString,
		Description:  data.Description,
		ThumbnailURL: data.ThumbnailURL,
		Visibility:   *data.Visibility,
	}

	if livestreamData.Title == "" {
		livestreamData.Title = "Livestream - " + time.Now().Format(time.RFC3339)
	}

	createdLivestream, err := s.livestreamRepo.Create(ctx, livestreamData)
	if err != nil {
		return nil, err
	}

	return createdLivestream, nil
}
