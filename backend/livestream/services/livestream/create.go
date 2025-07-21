package livestream

import (
	"context"
	"net/http"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/dto"
	serviceresponse "sen1or/letslive/livestream/responses"
	"sen1or/letslive/livestream/utils"
	"time"
)

func (s *LivestreamService) Create(ctx context.Context, data dto.CreateLivestreamRequestDTO) (*domains.Livestream, *serviceresponse.ServiceErrorResponse) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, serviceresponse.NewServiceErrorResponse(http.StatusBadRequest, err.Error())
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
