package services

import (
	"context"
	"mime/multipart"
	"sen1or/lets-live/user/domains"
	"sen1or/lets-live/user/dto"
	servererrors "sen1or/lets-live/user/errors"
	livestreamgateway "sen1or/lets-live/user/gateway/livestream"
	"sen1or/lets-live/user/mapper"
	"sen1or/lets-live/user/repositories"
	"sen1or/lets-live/user/utils"
	"strconv"

	"github.com/gofrs/uuid/v5"
)

type UserService struct {
	repo                      repositories.UserRepository
	livestreamInformationRepo repositories.LivestreamInformationRepository
	livestreamGateway         livestreamgateway.LivestreamGateway
	minioService              MinIOService
}

func NewUserService(
	repo repositories.UserRepository,
	livestreamInformationRepo repositories.LivestreamInformationRepository,
	livestreamGateway livestreamgateway.LivestreamGateway,
	minioService MinIOService,
) *UserService {
	return &UserService{
		repo:                      repo,
		livestreamInformationRepo: livestreamInformationRepo,
		livestreamGateway:         livestreamGateway,
		minioService:              minioService,
	}
}

func (s *UserService) GetUserById(userUUID uuid.UUID) (*dto.GetUserResponseDTO, *servererrors.ServerError) {
	user, err := s.repo.GetById(userUUID)
	if err != nil {
		return nil, err
	}

	userVODs, errRes := s.livestreamGateway.GetUserLivestreams(context.Background(), userUUID.String())
	if errRes != nil {
		return nil, servererrors.NewServerError(errRes.StatusCode, errRes.Message)
	}

	res := mapper.UserToGetUserResponseDTO(*user, userVODs)
	return res, nil
}

func (s *UserService) GetUserByStreamAPIKey(key uuid.UUID) (*dto.GetUserResponseDTO, *servererrors.ServerError) {
	user, err := s.repo.GetByAPIKey(key)
	if err != nil {
		return nil, err
	}

	res := mapper.UserToGetUserResponseDTO(*user, nil)
	return res, nil
}

func (s *UserService) GetUserFullInformation(userUUID uuid.UUID) (*domains.User, *servererrors.ServerError) {
	user, err := s.repo.GetById(userUUID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) QueryUsers(isOnline, username, page string) ([]dto.GetUserResponseDTO, *servererrors.ServerError) {
	var pageNumber int
	if len(page) == 0 {
		pageNumber = 0
	} else {
		atoiNum, atoiErr := strconv.Atoi(page)
		if atoiErr != nil {
			return nil, servererrors.ErrInvalidInput
		}
		pageNumber = atoiNum
	}

	users, err := s.repo.Query(isOnline, username, pageNumber)
	if err != nil {
		return nil, err
	}

	var resUsers []dto.GetUserResponseDTO

	for _, user := range users {
		userVODs, errRes := s.livestreamGateway.GetUserLivestreams(context.Background(), user.Id.String())
		if errRes != nil {
			continue // what should be done?
		}

		resUsers = append(resUsers, *mapper.UserToGetUserResponseDTO(*user, userVODs))
	}

	return resUsers, nil
}

func (s *UserService) CreateNewUser(data dto.CreateUserRequestDTO) *servererrors.ServerError {
	if err := utils.Validator.Struct(&data); err != nil {
		return servererrors.ErrInvalidInput
	}

	err := s.repo.Create(data.Username, data.Email, data.IsVerified)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) UpdateUserVerified(userId uuid.UUID) *servererrors.ServerError {
	err := s.repo.UpdateUserVerified(userId)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) UpdateUser(data dto.UpdateUserRequestDTO) (*domains.User, *servererrors.ServerError) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, servererrors.ErrInvalidInput
	}

	updateData := mapper.UpdateUserRequestDTOToUser(data)
	updatedUser, err := s.repo.Update(updateData)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s *UserService) UpdateUserAPIKey(userId uuid.UUID) (string, *servererrors.ServerError) {
	newStreamKey, genErr := uuid.NewGen().NewV4()
	if genErr != nil {
		return "", servererrors.ErrInternalServer
	}

	err := s.repo.UpdateStreamAPIKey(userId, newStreamKey.String())
	if err != nil {
		return "", err
	}

	return newStreamKey.String(), nil
}

func (s UserService) UpdateUserProfilePicture(file multipart.File, fileHeader *multipart.FileHeader, userId uuid.UUID) (string, *servererrors.ServerError) {
	savedPath, err := s.minioService.AddFile(file, fileHeader, "profile-pictures")
	if err != nil {
		return "", servererrors.ErrInternalServer
	}

	updateErr := s.repo.UpdateProfilePicture(userId, savedPath)
	if updateErr != nil {
		return "", updateErr
	}

	return savedPath, nil
}

func (s UserService) UpdateUserBackgroundPicture(file multipart.File, fileHeader *multipart.FileHeader, userId uuid.UUID) (string, *servererrors.ServerError) {
	savedPath, err := s.minioService.AddFile(file, fileHeader, "background-pictures")
	if err != nil {
		return "", servererrors.ErrInternalServer
	}

	updateErr := s.repo.UpdateBackgroundPicture(userId, savedPath)
	if updateErr != nil {
		return "", updateErr
	}

	return savedPath, nil
}

// INTERNAL USE
func (s UserService) UpdateUserInternal(data dto.UpdateUserRequestDTO) (*domains.User, *servererrors.ServerError) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, servererrors.ErrInvalidInput
	}

	updateUser := mapper.UpdateUserRequestDTOToUser(data)

	updatedUser, err := s.repo.Update(updateUser)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}
