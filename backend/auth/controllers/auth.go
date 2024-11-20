package controllers

import (
	"sen1or/lets-live/auth/domains"
	"sen1or/lets-live/auth/dto"
	"sen1or/lets-live/auth/mapper"
	"sen1or/lets-live/auth/repositories"

	"github.com/gofrs/uuid/v5"
)

type AuthController struct {
	repo repositories.AuthRepository
}

func NewAuthController(repo repositories.AuthRepository) *AuthController {
	return &AuthController{
		repo: repo,
	}
}

func (c *AuthController) Create(body domains.Auth) (*dto.SignUpResponseDTO, error) {
	createdAuth, err := c.repo.Create(body)
	if err != nil {
		return nil, err
	}

	return mapper.AuthToSignUpResponseDTO(*createdAuth), nil
}

func (c *AuthController) GetByID(id uuid.UUID) (*domains.Auth, error) {
	user, err := c.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *AuthController) GetByEmail(email string) (*domains.Auth, error) {
	user, err := c.repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *AuthController) UpdatePasswordHash(auth domains.Auth) (*domains.Auth, error) {
	updatedAuth, err := c.repo.UpdatePasswordHash(auth)
	return updatedAuth, err
}

func (c *AuthController) UpdateUserVerify(auth domains.Auth) (*domains.Auth, error) {
	updatedAuth, err := c.repo.UpdateVerify(auth)
	return updatedAuth, err
}

//
//func (c *AuthController) Delete(userID uuid.UUID) error {
//	return c.repo.Delete(userID)
//}
