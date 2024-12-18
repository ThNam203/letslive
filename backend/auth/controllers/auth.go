package controllers

import (
	"sen1or/lets-live/auth/domains"
	"sen1or/lets-live/auth/dto"
	"sen1or/lets-live/auth/mapper"
	"sen1or/lets-live/auth/repositories"

	"github.com/gofrs/uuid/v5"
)

type AuthController interface {
	Create(body domains.Auth) (*dto.SignUpResponseDTO, error)
	GetByID(id uuid.UUID) (*domains.Auth, error)
	GetByEmail(email string) (*domains.Auth, error)
	UpdatePasswordHash(auth domains.Auth) (*domains.Auth, error)
	UpdateUserVerify(auth domains.Auth) (*domains.Auth, error)
}

type authController struct {
	repo repositories.AuthRepository
}

func NewAuthController(repo repositories.AuthRepository) AuthController {
	return &authController{
		repo: repo,
	}
}

func (c *authController) Create(body domains.Auth) (*dto.SignUpResponseDTO, error) {
	createdAuth, err := c.repo.Create(body)
	if err != nil {
		return nil, err
	}

	return mapper.AuthToSignUpResponseDTO(*createdAuth), nil
}

func (c *authController) GetByID(id uuid.UUID) (*domains.Auth, error) {
	user, err := c.repo.GetByID(id)
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

func (c *authController) UpdateUserVerify(auth domains.Auth) (*domains.Auth, error) {
	updatedAuth, err := c.repo.UpdateVerify(auth)
	return updatedAuth, err
}

//
//func (c *AuthController) Delete(userID uuid.UUID) error {
//	return c.repo.Delete(userID)
//}
