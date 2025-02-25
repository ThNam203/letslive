package controllers

import (
	"sen1or/lets-live/livestream/domains"
	"sen1or/lets-live/livestream/dto"
	"sen1or/lets-live/livestream/mapper"
	"sen1or/lets-live/livestream/repositories"

	"github.com/gofrs/uuid/v5"
)

type LivestreamController interface {
	GetById(uuid.UUID) (*domains.Livestream, error)
	GetByUser(uuid.UUID) ([]domains.Livestream, error)

	Create(dto.CreateLivestreamRequestDTO) (*domains.Livestream, error)
	Update(dto.UpdateLivestreamRequestDTO) (*domains.Livestream, error)
	Delete(uuid.UUID) error
}

type livestreamController struct {
	repo repositories.LivestreamRepository
}

func NewLivestreamController(repo repositories.LivestreamRepository) LivestreamController {
	return &livestreamController{
		repo: repo,
	}
}

func (c *livestreamController) GetById(livestreamId uuid.UUID) (*domains.Livestream, error) {
	return c.repo.GetById(livestreamId)
}

func (c *livestreamController) GetByUser(userId uuid.UUID) ([]domains.Livestream, error) {
	return c.repo.GetByUser(userId)
}

func (c *livestreamController) Create(dto dto.CreateLivestreamRequestDTO) (*domains.Livestream, error) {
	livestreamData := mapper.CreateLivestreamRequestDTOToLivestream(dto)
	createdLivestream, err := c.repo.Create(livestreamData)
	if err != nil {
		return nil, err
	}

	return createdLivestream, nil
}

func (c *livestreamController) Update(dto dto.UpdateLivestreamRequestDTO) (*domains.Livestream, error) {
	currentLivestream, err := c.repo.GetById(dto.Id)
	if err != nil {
		return nil, err
	}

	// TODO: is this the best way?
	if dto.Title != nil {
		currentLivestream.Title = dto.Title
	}
	if dto.Description != nil {
		currentLivestream.Description = dto.Description
	}
	if dto.ThumbnailURL != nil {
		currentLivestream.ThumbnailURL = dto.ThumbnailURL
	}
	if dto.Status != nil {
		currentLivestream.Status = *dto.Status
	}
	if dto.PlaybackURL != nil {
		currentLivestream.PlaybackURL = dto.PlaybackURL
	}
	if dto.ViewCount != nil {
		currentLivestream.ViewCount = *dto.ViewCount
	}
	if dto.EndedAt != nil {
		currentLivestream.EndedAt = dto.EndedAt
	}

	updatedLivestream, err := c.repo.Update(*currentLivestream)
	if err != nil {
		return nil, err
	}

	return updatedLivestream, nil
}

func (c *livestreamController) Delete(livestreamId uuid.UUID) error {
	return c.repo.Delete(livestreamId)
}
