package controllers

import (
	"sen1or/lets-live/user/dto"
	"sen1or/lets-live/user/mapper"
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

func (c *UserController) Create(body dto.CreateUserRequestDTO) (*dto.CreateUserResponseDTO, error) {
	user := mapper.CreateUserRequestDTOToUser(body)
	createdUser, err := c.repo.Create(*user)
	if err != nil {
		return nil, err
	}

	return mapper.UserToCreateUserResponseDTO(*createdUser), nil
}

func (c *UserController) GetByID(id uuid.UUID) (*dto.GetUserResponseDTO, error) {
	user, err := c.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return mapper.UserToGetUserResponseDTO(*user), nil
}

func (c *UserController) GetByEmail(email string) (*dto.GetUserResponseDTO, error) {
	user, err := c.repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	return mapper.UserToGetUserResponseDTO(*user), nil
}

func (c *UserController) Update(updateDTO dto.UpdateUserRequestDTO) (*dto.UpdateUserResponseDTO, error) {
	user := mapper.UpdateUserRequestDTOToUser(updateDTO)
	updatedUser, err := c.repo.Update(*user)

	if err != nil {
		return nil, err
	}

	return mapper.UserToUpdateUserResponseDTO(*updatedUser), nil
}

func (c *UserController) Delete(userID uuid.UUID) error {
	return c.repo.Delete(userID)
}
