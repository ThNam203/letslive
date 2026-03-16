package livestream

import (
	"context"
	"sen1or/letslive/livestream/dto"
	vodgateway "sen1or/letslive/livestream/gateway/vod"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"
	"sen1or/letslive/livestream/utils"
	"time"

	"github.com/gofrs/uuid/v5"
)

func (s *LivestreamService) EndLivestreamAndCreateVOD(ctx context.Context, streamId uuid.UUID, endReqDTO dto.EndLivestreamRequestDTO) *response.Response[any] {
	if err := utils.Validator.Struct(&endReqDTO); err != nil {
		return response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil)
	}

	currentLivestream, err := s.livestreamRepo.GetById(ctx, streamId)
	if err != nil {
		return err
	}

	if currentLivestream.EndedAt != nil {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_END_ALREADY_ENDED_LIVESTREAM,
			nil,
			nil,
			nil,
		)
	}

	now := time.Now()
	currentLivestream.EndedAt = &now
	updatedLs, err := s.livestreamRepo.Update(ctx, *currentLivestream)
	if err != nil {
		return err
	}

	// Create VOD via VOD service gateway
	var description string
	if currentLivestream.Description != nil {
		description = *currentLivestream.Description
	}
	var thumbnailURL string
	if currentLivestream.ThumbnailURL != nil {
		thumbnailURL = *currentLivestream.ThumbnailURL
	}
	var playbackURL string
	if endReqDTO.PlaybackURL != nil {
		playbackURL = *endReqDTO.PlaybackURL
	}

	createReq := vodgateway.CreateVODRequest{
		LivestreamId: currentLivestream.Id.String(),
		UserId:       currentLivestream.UserId.String(),
		Title:        currentLivestream.Title,
		Description:  description,
		ThumbnailURL: thumbnailURL,
		PlaybackURL:  playbackURL,
		Duration:     endReqDTO.Duration,
	}

	vodId, createErr := s.vodGateway.CreateVOD(ctx, createReq)
	if createErr != nil {
		logger.Warnf(ctx, "failed to create VOD via gateway for livestream %s: %v", currentLivestream.Id, createErr)
		return response.NewResponseFromTemplate[any](response.RES_ERR_VOD_CREATE_FAILED, nil, nil, nil)
	}

	if vodId != nil {
		updatedLs.VODId = vodId
	}
	updatedLs.EndedAt = endReqDTO.EndedAt
	_, updateErr := s.livestreamRepo.Update(ctx, *updatedLs)
	if updateErr != nil {
		logger.Warnf(ctx, "failed to link VOD id %s to livestream id %s: %v", vodId, updatedLs.Id, updateErr)
	}

	return nil
}
