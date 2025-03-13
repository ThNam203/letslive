package services

import (
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/dto"
	servererrors "sen1or/letslive/livestream/errors"
	"sen1or/letslive/livestream/mapper"
	"sen1or/letslive/livestream/repositories"
	"sen1or/letslive/livestream/utils"
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

func (c LivestreamService) GetAllLivestreaming(page int) ([]domains.Livestream, *servererrors.ServerError) {
	if page < 0 {
		return nil, servererrors.ErrInvalidInput
	}

	return c.repo.GetAllLivestreamings(page)
}

func (c LivestreamService) GetByUserPublic(userId uuid.UUID) ([]domains.Livestream, *servererrors.ServerError) {
	livestreams, err := c.repo.GetByUser(userId)
	if err != nil {
		return nil, err
	}

	publicLivestreams := []domains.Livestream{}
	for _, l := range livestreams {
		if l.Visibility == domains.LivestreamPublicVisibility {
			publicLivestreams = append(publicLivestreams, l)
		}
	}

	return publicLivestreams, nil
}

func (c LivestreamService) GetByUserAuthor(userId uuid.UUID) ([]domains.Livestream, *servererrors.ServerError) {
	return c.repo.GetByUser(userId)
}

func (c LivestreamService) GetPopularVODs(page int) ([]domains.Livestream, *servererrors.ServerError) {
	return c.repo.GetPopularVODs(page)
}

func (c LivestreamService) CheckIsUserLivestreaming(userId uuid.UUID) (bool, *servererrors.ServerError) {
	return c.repo.CheckIsUserLivestreaming(userId)
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

// if authorId is null, it means the transcode service is updating
func (c LivestreamService) Update(data dto.UpdateLivestreamRequestDTO, streamId uuid.UUID, authorId *uuid.UUID) (*domains.Livestream, *servererrors.ServerError) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, servererrors.ErrInvalidInput
	}

	currentLivestream, err := c.repo.GetById(streamId)
	if err != nil {
		return nil, err
	}

	if authorId != nil && *authorId != currentLivestream.UserId {
		return nil, servererrors.ErrForbidden
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
	if data.Visibility != nil {
		currentLivestream.Visibility = domains.LivestreamVisibility(*data.Visibility)
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
	if data.Duration != nil {
		currentLivestream.Duration = data.Duration
	}

	updatedLivestream, err := c.repo.Update(*currentLivestream)
	if err != nil {
		return nil, err
	}

	return updatedLivestream, nil
}

func (c LivestreamService) Delete(livestreamId, userId uuid.UUID) *servererrors.ServerError {
	livestream, err := c.repo.GetById(livestreamId)
	if err != nil {
		return err
	}

	if livestream.UserId != userId {
		return servererrors.ErrForbidden
	}

	return c.repo.Delete(livestreamId)
}
