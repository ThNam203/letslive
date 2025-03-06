package services

import (
	"sen1or/lets-live/livestream/domains"
	"sen1or/lets-live/livestream/dto"
	servererrors "sen1or/lets-live/livestream/errors"
	"sen1or/lets-live/livestream/mapper"
	"sen1or/lets-live/livestream/repositories"
	"sen1or/lets-live/livestream/utils"
	"time"

	"github.com/gofrs/uuid/v5"
)

type LivestreamService struct {
	repo repositories.LivestreamRepository
}

func NewLivestreamService(repo repositories.LivestreamRepository) *LivestreamService {
	return &LivestreamService{
		repo: repo,
	}
}

func (c LivestreamService) GetById(livestreamId uuid.UUID) (*domains.Livestream, *servererrors.ServerError) {
	return c.repo.GetById(livestreamId)
}

func (c LivestreamService) GetByUser(userId uuid.UUID) ([]domains.Livestream, *servererrors.ServerError) {
	return c.repo.GetByUser(userId)
}

func (c LivestreamService) Create(data dto.CreateLivestreamRequestDTO) (*domains.Livestream, *servererrors.ServerError) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, servererrors.ErrInvalidInput
	}

	if data.Title == nil {
		newTitle := "Livestream - " + time.Now().Format(time.RFC3339)
		data.Title = &newTitle
	}

	livestreamData := mapper.CreateLivestreamRequestDTOToLivestream(data)
	createdLivestream, err := c.repo.Create(livestreamData)
	if err != nil {
		return nil, err
	}

	return createdLivestream, nil
}

func (c LivestreamService) Update(data dto.UpdateLivestreamRequestDTO, streamId uuid.UUID) (*domains.Livestream, *servererrors.ServerError) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, servererrors.ErrInvalidInput
	}

	currentLivestream, err := c.repo.GetById(streamId)
	if err != nil {
		return nil, err
	}

	// TODO: is this the best way?
	if data.Title != nil {
		currentLivestream.Title = data.Title
	}
	if data.Description != nil {
		currentLivestream.Description = data.Description
	}
	if data.ThumbnailURL != nil {
		currentLivestream.ThumbnailURL = data.ThumbnailURL
	}
	if data.Status != nil {
		currentLivestream.Status = *data.Status
	}
	if data.PlaybackURL != nil {
		currentLivestream.PlaybackURL = data.PlaybackURL
	}
	if data.ViewCount != nil {
		currentLivestream.ViewCount = *data.ViewCount
	}
	if data.EndedAt != nil {
		currentLivestream.EndedAt = data.EndedAt
	}

	updatedLivestream, err := c.repo.Update(*currentLivestream)
	if err != nil {
		return nil, err
	}

	return updatedLivestream, nil
}

func (c LivestreamService) Delete(livestreamId uuid.UUID) *servererrors.ServerError {
	return c.repo.Delete(livestreamId)
}
