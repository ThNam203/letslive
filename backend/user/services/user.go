package services

import (
	"sen1or/lets-live/user/domains"
	"sen1or/lets-live/user/dto"
	"sen1or/lets-live/user/mapper"
	"sen1or/lets-live/user/repositories"

	"github.com/gofrs/uuid/v5"
)

// TODO: refactor the job of controller (i think the handler should handle parsing and validate data while the controller deal with logics, right now the handler is doing all the work)
type UserController interface {
	Create(body dto.CreateUserRequestDTO) (*domains.User, error)
	GetAll() ([]*dto.GetUserResponseDTO, error)
	GetById(id uuid.UUID) (*dto.GetUserResponseDTO, error)
	GetUserFullProfile(id uuid.UUID) (*domains.User, error)
	GetByEmail(email string) (*dto.GetUserResponseDTO, error)
	GetByStreamAPIKey(key uuid.UUID) (*domains.User, error)
	GetStreamingUsers() ([]*dto.GetUserResponseDTO, error)
	Update(updateDTO dto.UpdateUserRequestDTO) (*domains.User, error)
	UpdateStreamAPIKey(userId uuid.UUID) (string, error)
	UpdateUserVerified(userId uuid.UUID) error
	UpdateProfilePicture(id uuid.UUID, picturePath string) error
	UpdateBackgroundPicture(id uuid.UUID, picturePath string) error
	Delete(userID uuid.UUID) error
}

type userController struct {
	repo                      repositories.UserRepository
	livestreamInformationRepo repositories.LivestreamInformationRepository
}

func NewUserController(repo repositories.UserRepository, livestreamInformationController repositories.LivestreamInformationRepository) UserController {
	return &userController{
		repo:                      repo,
		livestreamInformationRepo: livestreamInformationController,
	}
}

// TODO: transaction please
func (c *userController) Create(body dto.CreateUserRequestDTO) (*domains.User, error) {
	user := mapper.CreateUserRequestDTOToUser(body)
	createdUser, err := c.repo.Create(*user)
	if err != nil {
		return nil, err
	}

	_ = c.livestreamInformationRepo.Create(createdUser.Id)

	return createdUser, nil
}

func (c *userController) GetById(id uuid.UUID) (*dto.GetUserResponseDTO, error) {
	user, err := c.repo.GetById(id)
	if err != nil {
		return nil, err
	}

	return mapper.UserToGetUserResponseDTO(*user), nil
}

func (c *userController) GetUserFullProfile(id uuid.UUID) (*domains.User, error) {
	user, err := c.repo.GetById(id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *userController) GetAll() ([]*dto.GetUserResponseDTO, error) {
	users, err := c.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var res = []*dto.GetUserResponseDTO{}
	for _, u := range users {
		res = append(res, mapper.UserToGetUserResponseDTO(*u))
	}
	return res, nil
}

func (c *userController) GetByEmail(email string) (*dto.GetUserResponseDTO, error) {
	user, err := c.repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	return mapper.UserToGetUserResponseDTO(*user), nil
}

func (c *userController) GetByStreamAPIKey(key uuid.UUID) (*domains.User, error) {
	user, err := c.repo.GetByAPIKey(key)
	if err != nil {
		return nil, err
	}

	return user, nil
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

func (c *userController) Update(updateDTO dto.UpdateUserRequestDTO) (*domains.User, error) {
	updateUser := mapper.UpdateUserRequestDTOToUser(updateDTO)
	updatedUser, err := c.repo.Update(updateUser)

	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (c *userController) UpdateStreamAPIKey(userId uuid.UUID) (string, error) {
	newStreamKey, err := uuid.NewGen().NewV4()
	if err != nil {
		return "", err
	}

	err = c.repo.UpdateStreamAPIKey(userId, newStreamKey.String())
	if err != nil {
		return "", err
	}

	return newStreamKey.String(), nil
}

func (c *userController) Delete(userID uuid.UUID) error {
	return c.repo.Delete(userID)
}

func (c *userController) UpdateProfilePicture(id uuid.UUID, picturePath string) error {
	return c.repo.UpdateProfilePicture(id, picturePath)
}

func (c *userController) UpdateBackgroundPicture(id uuid.UUID, picturePath string) error {
	return c.repo.UpdateBackgroundPicture(id, picturePath)
}

func (c *userController) UpdateUserVerified(userId uuid.UUID) error {
	return c.repo.UpdateUserVerified(userId)
}
