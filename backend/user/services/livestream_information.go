package services

import (
	"sen1or/lets-live/user/domains"
	servererrors "sen1or/lets-live/user/errors"
	"sen1or/lets-live/user/repositories"

	"github.com/gofrs/uuid/v5"
)

type LivestreamInformationService struct {
	repo repositories.LivestreamInformationRepository
}

func NewLivestreamInformationService(repo repositories.LivestreamInformationRepository) *LivestreamInformationService {
	return &LivestreamInformationService{
		repo: repo,
	}
}

func (c *LivestreamInformationService) Create(userId uuid.UUID) error {
	return c.repo.Create(userId)
}

func (c *LivestreamInformationService) Update(data domains.LivestreamInformation) (*domains.LivestreamInformation, *servererrors.ServerError) {
	updatedInformation, err := c.repo.Update(data)

	if err != nil {
		return nil, err
	}

	return updatedInformation, nil
}
