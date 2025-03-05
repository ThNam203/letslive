package services

import (
	"sen1or/lets-live/user/domains"
	"sen1or/lets-live/user/repositories"

	"github.com/gofrs/uuid/v5"
)

// TODO: refactor the job of controller (i think the handler should handle parsing and validate data while the controller deal with logics, right now the handler is doing all the work)
type LivestreamInformationController interface {
	Create(userId uuid.UUID) error
	Update(data domains.LivestreamInformation) (*domains.LivestreamInformation, error)
}

type livestreamInformationController struct {
	repo repositories.LivestreamInformationRepository
}

func NewLivestreamInformationController(repo repositories.LivestreamInformationRepository) LivestreamInformationController {
	return &livestreamInformationController{
		repo: repo,
	}
}

func (c *livestreamInformationController) Create(userId uuid.UUID) error {
	return c.repo.Create(userId)
}

func (c *livestreamInformationController) Update(data domains.LivestreamInformation) (*domains.LivestreamInformation, error) {
	updatedInformation, err := c.repo.Update(data)

	if err != nil {
		return nil, err
	}

	return updatedInformation, nil
}
