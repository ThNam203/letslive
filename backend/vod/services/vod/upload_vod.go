package vod

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sen1or/letslive/vod/domains"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/response"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
)

var allowedVideoExtensions = map[string]bool{
	".mp4":  true,
	".mov":  true,
	".avi":  true,
	".mkv":  true,
	".webm": true,
}

func (s *VODService) UploadVOD(
	ctx context.Context,
	userId uuid.UUID,
	title string,
	description string,
	visibility string,
	filename string,
	fileSize int64,
	fileReader io.Reader,
) (*domains.VOD, *response.Response[any]) {
	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedVideoExtensions[ext] {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		)
	}

	// Generate VOD ID upfront for the raw file path
	vodId, err := uuid.NewV4()
	if err != nil {
		logger.Errorf(ctx, "failed to generate uuid: %v", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	// Upload raw file to MinIO
	objectName := fmt.Sprintf("raw-videos/%s/%s", vodId.String(), filename)
	rawFileURL, uploadErr := s.minioStorage.UploadFile(ctx, objectName, fileReader, fileSize, "video/"+strings.TrimPrefix(ext, "."))
	if uploadErr != nil {
		logger.Errorf(ctx, "failed to upload raw video to minio: %v", uploadErr)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	// Determine visibility
	vodVisibility := domains.VODPublicVisibility
	if visibility == "private" {
		vodVisibility = domains.VODPrivateVisibility
	}

	now := time.Now()
	desc := &description
	vodData := domains.VOD{
		Id:              vodId,
		UserId:          userId,
		Title:           title,
		Description:     desc,
		Visibility:      vodVisibility,
		Status:          domains.VODStatusProcessing,
		OriginalFileURL: &rawFileURL,
		ViewCount:       0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	createdVOD, createErr := s.vodRepo.Create(ctx, vodData)
	if createErr != nil {
		return nil, createErr
	}

	// Create transcode job
	transcodeJob := domains.TranscodeJob{
		VodId:       createdVOD.Id,
		Status:      domains.TranscodeJobPending,
		Attempts:    0,
		MaxAttempts: 3,
	}

	_, jobErr := s.transcodeJobRepo.Create(ctx, transcodeJob)
	if jobErr != nil {
		logger.Errorf(ctx, "failed to create transcode job for vod %s: %v", createdVOD.Id, jobErr)
		// VOD is created but job failed - mark VOD as failed
		s.vodRepo.UpdateStatus(ctx, createdVOD.Id, domains.VODStatusFailed, nil, nil)
		return nil, jobErr
	}

	return createdVOD, nil
}

func (s *VODService) UpdateVODStatus(
	ctx context.Context,
	vodId uuid.UUID,
	status domains.VODStatus,
	playbackUrl *string,
	thumbnailUrl *string,
) *response.Response[any] {
	return s.vodRepo.UpdateStatus(ctx, vodId, status, playbackUrl, thumbnailUrl)
}
