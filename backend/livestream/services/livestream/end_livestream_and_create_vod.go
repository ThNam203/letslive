package livestream

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/dto"
	"sen1or/letslive/livestream/pkg/logger"
	serviceresponse "sen1or/letslive/livestream/responses"
	"time"

	"github.com/gofrs/uuid/v5"
)

func (s *LivestreamService) EndLivestreamAndCreateVOD(ctx context.Context, streamId uuid.UUID, endReqDTO dto.EndLivestreamRequestDTO) *serviceresponse.ServiceErrorResponse {
	currentLivestream, err := s.livestreamRepo.GetById(ctx, streamId)
	if err != nil {
		return err
	}

	if currentLivestream.EndedAt != nil {
		return serviceresponse.ErrEndAnAlreadyEndedLivestream
	}

	now := time.Now()
	currentLivestream.EndedAt = &now
	updatedLs, err := s.livestreamRepo.Update(ctx, *currentLivestream)
	if err != nil {
		return err
	}

	vodData := &domains.VOD{
		LivestreamId: currentLivestream.Id,
		UserId:       currentLivestream.UserId,
		Title:        &currentLivestream.Title,
		Description:  currentLivestream.Description,
		ThumbnailURL: currentLivestream.ThumbnailURL,
		Visibility:   domains.VODVisibility(currentLivestream.Visibility),
		ViewCount:    0,
		Duration:     &endReqDTO.Duration,
		PlaybackURL:  endReqDTO.PlaybackURL,
		UpdatedAt:    now,
	}

	createdVOD, err := s.vodRepo.Create(ctx, *vodData)
	if err != nil {
		// TODO: What happens if VOD creation fails after stream is marked ended? maybe a background job queue?
		return err
	}

	updatedLs.VODId = &createdVOD.Id
	updatedLs.EndedAt = endReqDTO.EndedAt
	_, updateErr := s.livestreamRepo.Update(ctx, *updatedLs)
	if updateErr != nil {
		logger.Warnf("failed to link VOD id %s to livestream id %s: %v", createdVOD.Id, updatedLs.Id, updateErr)
	}

	return nil
}
