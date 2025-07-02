package livestream

import (
	"context"
	"net/http"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/dto"
	serviceresponse "sen1or/letslive/livestream/responses"
	"sen1or/letslive/livestream/utils"
	"time"

	"github.com/gofrs/uuid/v5"
)

func (s *LivestreamService) Create(ctx context.Context, data dto.CreateLivestreamRequestDTO, userId uuid.UUID) (*domains.Livestream, *serviceresponse.ServiceErrorResponse) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, serviceresponse.NewServiceErrorResponse(http.StatusBadRequest, err.Error())
	}

	livestreamData := domains.Livestream{
		UserId:       data.UserId,
		Title:        *data.Title,
		Description:  data.Description,
		ThumbnailURL: data.ThumbnailURL,
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
