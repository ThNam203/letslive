package services

import (
	"mime/multipart"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"
	servererrors "sen1or/letslive/user/errors"
	"sen1or/letslive/user/mapper"
	"sen1or/letslive/user/pkg/logger"
	"sen1or/letslive/user/repositories"
	"sen1or/letslive/user/utils"

	"github.com/gofrs/uuid/v5"
)

type UserService struct {
	userRepo                  repositories.UserRepository
	livestreamInformationRepo repositories.LivestreamInformationRepository
	minioService              MinIOService
}

func NewUserService(
	userRepo repositories.UserRepository,
	livestreamInformationRepo repositories.LivestreamInformationRepository,
	minioService MinIOService,
) *UserService {
	return &UserService{
		userRepo:                  userRepo,
		livestreamInformationRepo: livestreamInformationRepo,
		minioService:              minioService,
	}
}

func (s *UserService) GetUserById(userUUID uuid.UUID, authenticatedUserId *uuid.UUID) (*dto.GetUserPublicResponseDTO, *servererrors.ServerError) {
	user, err := s.userRepo.GetById(userUUID, authenticatedUserId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserByStreamAPIKey(key uuid.UUID) (*domains.User, *servererrors.ServerError) {
	user, err := s.userRepo.GetByAPIKey(key)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserFullInformation(userUUID uuid.UUID) (*domains.User, *servererrors.ServerError) {
	user, err := s.userRepo.GetFullInfoById(userUUID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetAllUsers(page int) ([]domains.User, *servererrors.ServerError) {
	users, err := s.userRepo.GetAll(page)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *UserService) SearchUserByUsername(username string) ([]dto.GetUserPublicResponseDTO, *servererrors.ServerError) {
	if len(username) == 0 {
		return nil, servererrors.ErrInvalidInput
	}

	users, err := s.userRepo.SearchUserByUsername(username)
	if err != nil {
		return nil, err
	}

	var resUsers []dto.GetUserPublicResponseDTO

	for _, user := range users {
		resUsers = append(resUsers, *mapper.UserToGetUserPublicResponseDTO(*user))
	}

	return resUsers, nil
}

func (s *UserService) CreateNewUser(data dto.CreateUserRequestDTO) (*domains.User, *servererrors.ServerError) {
	if err := utils.Validator.Struct(&data); err != nil {
		logger.Debugf("failed to validate user create resquest: %+v", data)
		return nil, servererrors.ErrInvalidInput
	}

	// TODO: transaction please
	createdUser, err := s.userRepo.Create(data.Username, data.Email, data.IsVerified, domains.AuthProvider(data.AuthProvider))
	if err != nil {
		return nil, err
	}

	if err := s.livestreamInformationRepo.Create(createdUser.Id); err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (s *UserService) UpdateUserVerified(userId uuid.UUID) *servererrors.ServerError {
	err := s.userRepo.UpdateUserVerified(userId)
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
	updatedUser, err := s.userRepo.Update(updateData)
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

	err := s.userRepo.UpdateStreamAPIKey(userId, newStreamKey.String())
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

	updateErr := s.userRepo.UpdateProfilePicture(userId, savedPath)
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

	updateErr := s.userRepo.UpdateBackgroundPicture(userId, savedPath)
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

	updatedUser, err := s.userRepo.Update(updateUser)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s UserService) UploadFileToMinIO(file multipart.File, fileHeader *multipart.FileHeader) (string, *servererrors.ServerError) {
	savedPath, err := s.minioService.AddFile(file, fileHeader, "general-files")
	if err != nil {
		return "", servererrors.ErrInternalServer
	}

	return savedPath, nil
}
