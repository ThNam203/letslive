package controllers

import (
	"sen1or/lets-live/user/domains"
	"sen1or/lets-live/user/repositories"

	"github.com/gofrs/uuid/v5"
)

type UserController struct {
	repo repositories.UserRepository
}

func NewUserController(repo repositories.UserRepository) *UserController {
	return &UserController{
		repo: repo,
	}
}

func (c *UserController) Create(body *domains.User) error {
	if err := c.repo.Create(body); err != nil {
		return err
	}

	return nil
}

func (c *UserController) GetByID(id uuid.UUID) (*domains.User, error) {
	user, err := c.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *UserController) GetByEmail(email string) (*domains.User, error) {
	user, err := c.repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *UserController) Update(user *domains.User) error {
	return c.repo.Update(user)

}

func (c *UserController) Delete(userID uuid.UUID) error {
	return c.repo.Delete(userID)
}
