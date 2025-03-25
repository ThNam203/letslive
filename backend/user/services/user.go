package services

import (
	"context"
	"mime/multipart"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"
	servererrors "sen1or/letslive/user/errors"
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

func (s *UserService) GetUserPublicInfoById(ctx context.Context, userUUID uuid.UUID, authenticatedUserId *uuid.UUID) (*dto.GetUserPublicResponseDTO, *servererrors.ServerError) {
	user, err := s.userRepo.GetPublicInfoById(ctx, userUUID, authenticatedUserId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserByStreamAPIKey(ctx context.Context, key uuid.UUID) (*domains.User, *servererrors.ServerError) {
	user, err := s.userRepo.GetByAPIKey(ctx, key)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserById(ctx context.Context, userUUID uuid.UUID) (*domains.User, *servererrors.ServerError) {
	user, err := s.userRepo.GetById(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetAllUsers(ctx context.Context, page int) ([]domains.User, *servererrors.ServerError) {
	users, err := s.userRepo.GetAll(ctx, page)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *UserService) SearchUsersByUsername(ctx context.Context, username string) ([]dto.GetUserPublicResponseDTO, *servererrors.ServerError) {
	if len(username) == 0 {
		return nil, servererrors.ErrInvalidInput
	}

	users, err := s.userRepo.SearchUsersByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *UserService) CreateNewUser(ctx context.Context, data dto.CreateUserRequestDTO) (*domains.User, *servererrors.ServerError) {
	if err := utils.Validator.Struct(&data); err != nil {
		logger.Debugf("failed to validate user create resquest: %+v", data)
		return nil, servererrors.ErrInvalidInput
	}

	// TODO: transaction please
	createdUser, err := s.userRepo.Create(ctx, data.Username, data.Email, data.IsVerified, domains.AuthProvider(data.AuthProvider))
	if err != nil {
		return nil, err
	}

	if err := s.livestreamInformationRepo.Create(createdUser.Id); err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (s *UserService) UpdateUserVerified(ctx context.Context, userId uuid.UUID) *servererrors.ServerError {
	err := s.userRepo.UpdateUserVerified(ctx, userId)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) UpdateUser(ctx context.Context, data dto.UpdateUserRequestDTO) (*domains.User, *servererrors.ServerError) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, servererrors.ErrInvalidInput
	}

	updatedUser, err := s.userRepo.Update(ctx, data)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s *UserService) UpdateUserAPIKey(ctx context.Context, userId uuid.UUID) (string, *servererrors.ServerError) {
	newStreamKey, genErr := uuid.NewGen().NewV4()
	if genErr != nil {
		return "", servererrors.ErrInternalServer
	}

	err := s.userRepo.UpdateStreamAPIKey(ctx, userId, newStreamKey.String())
	if err != nil {
		return "", err
	}

	return newStreamKey.String(), nil
}

func (s UserService) UpdateUserProfilePicture(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, userId uuid.UUID) (string, *servererrors.ServerError) {
	savedPath, err := s.minioService.AddFile(ctx, file, fileHeader, "profile-pictures")
	if err != nil {
		return "", servererrors.ErrInternalServer
	}

	updateErr := s.userRepo.UpdateProfilePicture(ctx, userId, savedPath)
	if updateErr != nil {
		return "", updateErr
	}

	return savedPath, nil
}

func (s UserService) UpdateUserBackgroundPicture(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, userId uuid.UUID) (string, *servererrors.ServerError) {
	savedPath, err := s.minioService.AddFile(ctx, file, fileHeader, "background-pictures")
	if err != nil {
		return "", servererrors.ErrInternalServer
	}

	updateErr := s.userRepo.UpdateBackgroundPicture(ctx, userId, savedPath)
	if updateErr != nil {
		return "", updateErr
	}

	return savedPath, nil
}

// INTERNAL USE
func (s UserService) UpdateUserInternal(ctx context.Context, data dto.UpdateUserRequestDTO) (*domains.User, *servererrors.ServerError) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, servererrors.ErrInvalidInput
	}

	updatedUser, err := s.userRepo.Update(ctx, data)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s UserService) UploadFileToMinIO(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, *servererrors.ServerError) {
	savedPath, err := s.minioService.AddFile(ctx, file, fileHeader, "general-files")
	if err != nil {
		return "", servererrors.ErrInternalServer
	}

	return savedPath, nil
}
