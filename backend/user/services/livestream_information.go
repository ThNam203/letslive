package services

import (
	"sen1or/letslive/user/domains"
	servererrors "sen1or/letslive/user/errors"
)

type LivestreamInformationService struct {
	repo domains.LivestreamInformationRepository
}

func NewLivestreamInformationService(repo domains.LivestreamInformationRepository) *LivestreamInformationService {
	return &LivestreamInformationService{
		repo: repo,
	}
}

func (c *LivestreamInformationService) Update(data domains.LivestreamInformation) (*domains.LivestreamInformation, *servererrors.ServerError) {
	updatedInformation, err := c.repo.Update(data)

	if err != nil {
		return nil, err
	}

	return updatedInformation, nil
}
