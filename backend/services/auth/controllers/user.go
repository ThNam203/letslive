package controllers

import (
	"sen1or/lets-live/auth/domains"
	"sen1or/lets-live/auth/repositories"

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

func (c *UserController) GetByFacebookID(facebookID string) (*domains.User, error) {
	user, err := c.repo.GetByFacebookID(facebookID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *UserController) GetByName(username string) (*domains.User, error) {
	user, err := c.repo.GetByName(username)
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
