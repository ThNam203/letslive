package services

import (
	"context"
	"mime/multipart"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/utils"

	"github.com/gofrs/uuid/v5"
)

type UserService struct {
	userRepo                  domains.UserRepository
	livestreamInformationRepo domains.LivestreamInformationRepository
	notificationRepo          domains.NotificationRepository
	followRepo                domains.FollowRepository
	minioService              MinIOService
}

func NewUserService(
	userRepo domains.UserRepository,
	livestreamInformationRepo domains.LivestreamInformationRepository,
	notificationRepo domains.NotificationRepository,
	followRepo domains.FollowRepository,
	minioService MinIOService,
) *UserService {
	return &UserService{
		userRepo:                  userRepo,
		livestreamInformationRepo: livestreamInformationRepo,
		notificationRepo:          notificationRepo,
		followRepo:                followRepo,
		minioService:              minioService,
	}
}

func (s *UserService) GetUserPublicInfoById(ctx context.Context, userUUID uuid.UUID, authenticatedUserId *uuid.UUID) (*dto.GetUserPublicResponseDTO, error) {
	user, err := s.userRepo.GetPublicInfoById(ctx, userUUID, authenticatedUserId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserByStreamAPIKey(ctx context.Context, key uuid.UUID) (*domains.User, error) {
	user, err := s.userRepo.GetByAPIKey(ctx, key)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserById(ctx context.Context, userUUID uuid.UUID) (*domains.User, error) {
	user, err := s.userRepo.GetById(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetIdentitiesByIds returns the minimal identity of each user for peer services.
// It is the authoritative answer to "what is this user called", so callers never
// have to take a name on trust from their own clients.
func (s *UserService) GetIdentitiesByIds(ctx context.Context, ids []uuid.UUID) ([]dto.UserIdentityInternalResponseDTO, error) {
	users, err := s.userRepo.GetPublicInfosByIds(ctx, ids, nil)
	if err != nil {
		return nil, err
	}

	identities := make([]dto.UserIdentityInternalResponseDTO, 0, len(users))
	for _, user := range users {
		identities = append(identities, dto.UserIdentityInternalResponseDTO{
			Id:             user.Id,
			Username:       user.Username,
			ProfilePicture: user.ProfilePicture,
		})
	}

	return identities, nil
}

func (s *UserService) GetFollowingUsers(ctx context.Context, authenticatedUserId uuid.UUID) ([]dto.GetUserPublicResponseDTO, error) {
	ids, err := s.followRepo.GetFollowedUserIds(ctx, authenticatedUserId)
	if err != nil {
		return nil, err
	}
	users, err := s.userRepo.GetPublicInfosByIds(ctx, ids, &authenticatedUserId)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *UserService) GetRecommendedUsers(ctx context.Context, authenticatedUserId *uuid.UUID, page int) ([]dto.GetUserPublicResponseDTO, error) {
	users, err := s.userRepo.GetRecommendedPublic(ctx, authenticatedUserId, page, 10)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *UserService) SearchUsersByUsername(ctx context.Context, username string, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, error) {
	if len(username) == 0 {
		return nil, domains.ErrInvalidInput
	}

	users, err := s.userRepo.SearchUsersByUsername(ctx, username, authenticatedUserId)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *UserService) CreateNewUser(ctx context.Context, data dto.CreateUserRequestDTO) (*domains.User, error) {
	if err := utils.Validator.Struct(&data); err != nil {
		logger.Debugf(ctx, "failed to validate user create resquest: %+v", data)
		return nil, domains.ErrInvalidInput
	}

	// TODO: transaction please
	createdUser, err := s.userRepo.Create(ctx, data.Username, data.Email, domains.AuthProvider(data.AuthProvider))
	if err != nil {
		return nil, err
	}

	if err := s.livestreamInformationRepo.Create(ctx, createdUser.Id); err != nil {
		return nil, err
	}

	// Create welcome notification for the new user (ignore if fails)
	welcomeNotif := domains.Notification{
		UserId:  createdUser.Id,
		Type:    "system", // TODO: add system notification type enum/default values
		Title:   "Welcome to LetsLive!",
		Message: "Thanks for signing up. We're glad to have you here. Start by exploring streams or going live yourself.",
	}
	if _, err := s.notificationRepo.Create(ctx, welcomeNotif); err != nil {
		logger.Warnf(ctx, "failed to create welcome notification for user %s: %v", createdUser.Id, err)
	}

	return createdUser, nil
}

func (s *UserService) UpdateUser(ctx context.Context, data dto.UpdateUserRequestDTO) (*domains.User, error) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, domains.ErrInvalidInput
	}

	existedData, err := s.userRepo.GetById(ctx, data.Id)
	if err != nil {
		logger.Errorf(ctx, "failed to get existedData for user id: %s", data.Id)
		return nil, domains.ErrUserNotFound
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

	if data.Username != nil {
		existedData.Username = *data.Username
	}

	if data.PhoneNumber != nil {
		existedData.PhoneNumber = data.PhoneNumber
	}

	if data.Locale != nil {
		existedData.Locale = data.Locale
	}

	finalDTO := dto.UpdateUserRequestDTO{
		Id:               existedData.Id,
		Status:           statusPtr,
		Username:         &existedData.Username,
		PhoneNumber:      existedData.PhoneNumber,
		Bio:              existedData.Bio,
		Locale:           existedData.Locale,
		SocialMediaLinks: data.SocialMediaLinks,
	}

	updatedUser, err := s.userRepo.Update(ctx, finalDTO)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s *UserService) UpdateUserAPIKey(ctx context.Context, userId uuid.UUID) (string, error) {
	newStreamKey, genErr := uuid.NewGen().NewV4()
	if genErr != nil {
		return "", domains.ErrInternal
	}

	err := s.userRepo.UpdateStreamAPIKey(ctx, userId, newStreamKey.String())
	if err != nil {
		return "", err
	}

	return newStreamKey.String(), nil
}

func (s UserService) UpdateUserProfilePicture(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, userId uuid.UUID) (string, error) {
	savedPath, err := s.minioService.AddFile(ctx, file, fileHeader, "profile-pictures")
	if err != nil {
		return "", domains.ErrInternal
	}

	updateErr := s.userRepo.UpdateProfilePicture(ctx, userId, savedPath)
	if updateErr != nil {
		return "", updateErr
	}

	return savedPath, nil
}

func (s UserService) UpdateUserBackgroundPicture(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, userId uuid.UUID) (string, error) {
	savedPath, err := s.minioService.AddFile(ctx, file, fileHeader, "background-pictures")
	if err != nil {
		return "", domains.ErrInternal
	}

	updateErr := s.userRepo.UpdateBackgroundPicture(ctx, userId, savedPath)
	if updateErr != nil {
		return "", updateErr
	}

	return savedPath, nil
}

// INTERNAL USE
func (s UserService) UpdateUserInternal(ctx context.Context, data dto.UpdateUserRequestDTO) (*domains.User, error) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, domains.ErrInvalidInput
	}

	updatedUser, err := s.userRepo.Update(ctx, data)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s UserService) UploadFileToMinIO(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	savedPath, err := s.minioService.AddFile(ctx, file, fileHeader, "general-files")
	if err != nil {
		return "", domains.ErrInternal
	}

	return savedPath, nil
}
