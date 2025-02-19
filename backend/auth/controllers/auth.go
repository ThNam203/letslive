package controllers

import (
	"sen1or/lets-live/auth/domains"
	"sen1or/lets-live/auth/repositories"

	"github.com/gofrs/uuid/v5"
)

type AuthController interface {
	Create(body domains.Auth) (*domains.Auth, error)
	GetByID(id uuid.UUID) (*domains.Auth, error)
	GetByUserID(id uuid.UUID) (*domains.Auth, error)
	GetByEmail(email string) (*domains.Auth, error)
	UpdatePasswordHash(auth domains.Auth) (*domains.Auth, error)
}

type authController struct {
	repo repositories.AuthRepository
}

func NewAuthController(repo repositories.AuthRepository) AuthController {
	return &authController{
		repo: repo,
	}
}

func (c *authController) Create(body domains.Auth) (*domains.Auth, error) {
	createdAuth, err := c.repo.Create(body)
	if err != nil {
		return nil, err
	}

	return createdAuth, nil
}

func (c *authController) GetByID(id uuid.UUID) (*domains.Auth, error) {
	user, err := c.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *authController) GetByUserID(id uuid.UUID) (*domains.Auth, error) {
	user, err := c.repo.GetByUserID(id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *authController) GetByEmail(email string) (*domains.Auth, error) {
	user, err := c.repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *authController) UpdatePasswordHash(auth domains.Auth) (*domains.Auth, error) {
	updatedAuth, err := c.repo.UpdatePasswordHash(auth)
	return updatedAuth, err
}

//
//func (c *AuthController) Delete(userID uuid.UUID) error {
//	return c.repo.Delete(userID)
//}
