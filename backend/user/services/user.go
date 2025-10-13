package services

import (
	"context"
	"mime/multipart"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/pkg/logger"
	"sen1or/letslive/user/response"
	"sen1or/letslive/user/utils"

	"github.com/gofrs/uuid/v5"
)

type UserService struct {
	userRepo                  domains.UserRepository
	livestreamInformationRepo domains.LivestreamInformationRepository
	minioService              MinIOService
}

func NewUserService(
	userRepo domains.UserRepository,
	livestreamInformationRepo domains.LivestreamInformationRepository,
	minioService MinIOService,
) *UserService {
	return &UserService{
		userRepo:                  userRepo,
		livestreamInformationRepo: livestreamInformationRepo,
		minioService:              minioService,
	}
}

func (s *UserService) GetUserPublicInfoById(ctx context.Context, userUUID uuid.UUID, authenticatedUserId *uuid.UUID) (*dto.GetUserPublicResponseDTO, *response.Response[any]) {
	user, err := s.userRepo.GetPublicInfoById(ctx, userUUID, authenticatedUserId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserByStreamAPIKey(ctx context.Context, key uuid.UUID) (*domains.User, *response.Response[any]) {
	user, err := s.userRepo.GetByAPIKey(ctx, key)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserById(ctx context.Context, userUUID uuid.UUID) (*domains.User, *response.Response[any]) {
	user, err := s.userRepo.GetById(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetAllUsers(ctx context.Context, page int) ([]domains.User, *response.Response[any]) {
	users, err := s.userRepo.GetAll(ctx, page)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *UserService) SearchUsersByUsername(ctx context.Context, username string, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, *response.Response[any]) {
	if len(username) == 0 {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		)
	}

	users, err := s.userRepo.SearchUsersByUsername(ctx, username, authenticatedUserId)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *UserService) CreateNewUser(ctx context.Context, data dto.CreateUserRequestDTO) (*domains.User, *response.Response[any]) {
	if err := utils.Validator.Struct(&data); err != nil {
		logger.Debugf(ctx, "failed to validate user create resquest: %+v", data)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		)
	}

	// TODO: transaction please
	createdUser, err := s.userRepo.Create(ctx, data.Username, data.Email, domains.AuthProvider(data.AuthProvider))
	if err != nil {
		return nil, err
	}

	if err := s.livestreamInformationRepo.Create(ctx, createdUser.Id); err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (s *UserService) UpdateUser(ctx context.Context, data dto.UpdateUserRequestDTO) (*domains.User, *response.Response[any]) {
	existedData, err := s.userRepo.GetById(ctx, data.Id)
	if err != nil {
		logger.Errorf(ctx, "failed to get existedData for user id: %s", data.Id)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_USER_NOT_FOUND,
			nil,
			nil,
			nil,
		)
	}

	if data.Bio != nil {
		existedData.Bio = data.Bio
	}

	var statusPtr *string
	if data.Status != nil {
		statusPtr = data.Status
	} else {
		s := string(existedData.Status)
		statusPtr = &s
	}

	if data.DisplayName != nil {
		existedData.DisplayName = data.DisplayName
	}

	// currently username is not changable
	//if data.Username != nil {
	//	existedData.Username = *data.Username
	//}

	if data.PhoneNumber != nil {
		existedData.PhoneNumber = data.PhoneNumber
	}

	finalDTO := dto.UpdateUserRequestDTO{
		Id:               existedData.Id,
		Status:           statusPtr,
		PhoneNumber:      existedData.PhoneNumber,
		Bio:              existedData.Bio,
		DisplayName:      existedData.DisplayName,
		SocialMediaLinks: data.SocialMediaLinks,
	}

	updatedUser, err := s.userRepo.Update(ctx, finalDTO)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s *UserService) UpdateUserAPIKey(ctx context.Context, userId uuid.UUID) (string, *response.Response[any]) {
	newStreamKey, genErr := uuid.NewGen().NewV4()
	if genErr != nil {
		return "", response.NewResponseFromTemplate[any](
			response.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	err := s.userRepo.UpdateStreamAPIKey(ctx, userId, newStreamKey.String())
	if err != nil {
		return "", err
	}

	return newStreamKey.String(), nil
}

func (s UserService) UpdateUserProfilePicture(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, userId uuid.UUID) (string, *response.Response[any]) {
	savedPath, err := s.minioService.AddFile(ctx, file, fileHeader, "profile-pictures")
	if err != nil {
		return "", response.NewResponseFromTemplate[any](
			response.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	updateErr := s.userRepo.UpdateProfilePicture(ctx, userId, savedPath)
	if updateErr != nil {
		return "", updateErr
	}

	return savedPath, nil
}

func (s UserService) UpdateUserBackgroundPicture(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, userId uuid.UUID) (string, *response.Response[any]) {
	savedPath, err := s.minioService.AddFile(ctx, file, fileHeader, "background-pictures")
	if err != nil {
		return "", response.NewResponseFromTemplate[any](
			response.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	updateErr := s.userRepo.UpdateBackgroundPicture(ctx, userId, savedPath)
	if updateErr != nil {
		return "", updateErr
	}

	return savedPath, nil
}

// INTERNAL USE
func (s UserService) UpdateUserInternal(ctx context.Context, data dto.UpdateUserRequestDTO) (*domains.User, *response.Response[any]) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		)
	}

	updatedUser, err := s.userRepo.Update(ctx, data)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s UserService) UploadFileToMinIO(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, *response.Response[any]) {
	savedPath, err := s.minioService.AddFile(ctx, file, fileHeader, "general-files")
	if err != nil {
		return "", response.NewResponseFromTemplate[any](
			response.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	return savedPath, nil
}
