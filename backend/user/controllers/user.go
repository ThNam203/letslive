package controllers

import (
	"sen1or/lets-live/user/dto"
	"sen1or/lets-live/user/mapper"
	"sen1or/lets-live/user/repositories"

	"github.com/gofrs/uuid/v5"
)

type UserController interface {
	Create(body dto.CreateUserRequestDTO) (*dto.CreateUserResponseDTO, error)
	GetById(id uuid.UUID) (*dto.GetUserResponseDTO, error)
	GetByEmail(email string) (*dto.GetUserResponseDTO, error)
	GetByStreamAPIKey(key uuid.UUID) (*dto.GetUserResponseDTO, error)
	GetStreamingUsers() ([]*dto.GetUserResponseDTO, error)
	Update(updateDTO dto.UpdateUserRequestDTO) (*dto.UpdateUserResponseDTO, error)
	Delete(userID uuid.UUID) error
}

type userController struct {
	repo repositories.UserRepository
}

func NewUserController(repo repositories.UserRepository) UserController {
	return &userController{
		repo: repo,
	}
}

func (c *userController) Create(body dto.CreateUserRequestDTO) (*dto.CreateUserResponseDTO, error) {
	user := mapper.CreateUserRequestDTOToUser(body)
	createdUser, err := c.repo.Create(*user)
	if err != nil {
		return nil, err
	}

	return mapper.UserToCreateUserResponseDTO(*createdUser), nil
}

func (c *userController) GetById(id uuid.UUID) (*dto.GetUserResponseDTO, error) {
	user, err := c.repo.GetById(id)
	if err != nil {
		return nil, err
	}

	return mapper.UserToGetUserResponseDTO(*user), nil
}

func (c *userController) GetByEmail(email string) (*dto.GetUserResponseDTO, error) {
	user, err := c.repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	return mapper.UserToGetUserResponseDTO(*user), nil
}

func (c *userController) GetByStreamAPIKey(key uuid.UUID) (*dto.GetUserResponseDTO, error) {
	user, err := c.repo.GetByAPIKey(key)
	if err != nil {
		return nil, err
	}

	return mapper.UserToGetUserResponseDTO(*user), nil
}

func (c *userController) GetStreamingUsers() ([]*dto.GetUserResponseDTO, error) {
	onlineUsers, err := c.repo.GetStreamingUsers()
	if err != nil {
		return nil, err
	}

	var result = []*dto.GetUserResponseDTO{}
	for _, user := range onlineUsers {
		result = append(result, mapper.UserToGetUserResponseDTO(user))
	}

	return result, nil
}

func (c *userController) Update(updateDTO dto.UpdateUserRequestDTO) (*dto.UpdateUserResponseDTO, error) {
	updateUser, err := c.repo.GetById(updateDTO.Id)
	if err != nil {
		return nil, err
	}

	if updateDTO.Username != nil {
		updateUser.Username = *updateDTO.Username
	}

	if updateDTO.IsOnline != nil {
		updateUser.IsOnline = *updateDTO.IsOnline
	}

	updatedUser, err := c.repo.Update(*updateUser)

	if err != nil {
		return nil, err
	}

	return mapper.UserToUpdateUserResponseDTO(*updatedUser), nil
}

func (c *userController) Delete(userID uuid.UUID) error {
	return c.repo.Delete(userID)
}
