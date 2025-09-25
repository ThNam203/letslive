package services

import (
	"context"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/response"
)

type LivestreamInformationService struct {
	repo domains.LivestreamInformationRepository
}

func NewLivestreamInformationService(repo domains.LivestreamInformationRepository) *LivestreamInformationService {
	return &LivestreamInformationService{
		repo: repo,
	}
}

func (c *LivestreamInformationService) Update(ctx context.Context, data domains.LivestreamInformation) (*domains.LivestreamInformation, *response.Response[any]) {
	updatedInformation, err := c.repo.Update(ctx, data)

	if err != nil {
		return nil, err
	}

	return updatedInformation, nil
}
