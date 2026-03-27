package services

import (
	"context"
	"mime/multipart"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/response"
	"sen1or/letslive/user/utils"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	dbPool                    *pgxpool.Pool
	userRepo                  domains.UserRepository
	livestreamInformationRepo domains.LivestreamInformationRepository
	notificationRepo          domains.NotificationRepository
	followRepo                domains.FollowRepository
	minioService              MinIOService
}

func NewUserService(
	dbPool *pgxpool.Pool,
	userRepo domains.UserRepository,
	livestreamInformationRepo domains.LivestreamInformationRepository,
	notificationRepo domains.NotificationRepository,
	followRepo domains.FollowRepository,
	minioService MinIOService,
) *UserService {
	return &UserService{
		dbPool:                    dbPool,
		userRepo:                  userRepo,
		livestreamInformationRepo: livestreamInformationRepo,
		notificationRepo:          notificationRepo,
		followRepo:                followRepo,
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

func (s *UserService) GetFollowingUsers(ctx context.Context, authenticatedUserId uuid.UUID) ([]dto.GetUserPublicResponseDTO, *response.Response[any]) {
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

func (s *UserService) GetRecommendedUsers(ctx context.Context, authenticatedUserId *uuid.UUID, page int) ([]dto.GetUserPublicResponseDTO, *response.Response[any]) {
	users, err := s.userRepo.GetRecommendedPublic(ctx, authenticatedUserId, page, 10)
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

	tx, txErr := s.dbPool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		logger.Errorf(ctx, "failed to begin transaction: %v", txErr)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}
	defer tx.Rollback(ctx)

	// Create user within transaction
	params := pgx.NamedArgs{
		"username":      data.Username,
		"email":         data.Email,
		"auth_provider": domains.AuthProvider(data.AuthProvider),
	}
	row, queryErr := tx.Query(ctx, "INSERT INTO users (username, email, auth_provider) VALUES (@username, @email, @auth_provider) RETURNING *", params)
	if queryErr != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}
	createdUser, collectErr := pgx.CollectOneRow(row, pgx.RowToStructByNameLax[domains.User])
	if collectErr != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	// Create livestream information within the same transaction
	result, execErr := tx.Exec(ctx, "INSERT INTO livestream_information (user_id) VALUES ($1)", createdUser.Id)
	if execErr != nil || result.RowsAffected() == 0 {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		logger.Errorf(ctx, "failed to commit transaction: %v", commitErr)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	// Create welcome notification outside the transaction (best-effort)
	welcomeNotif := domains.Notification{
		UserId:  createdUser.Id,
		Type:    "system", // TODO: add system notification type enum/default values
		Title:   "Welcome to LetsLive!",
		Message: "Thanks for signing up. We're glad to have you here. Start by exploring streams or going live yourself.",
	}
	if _, err := s.notificationRepo.Create(ctx, welcomeNotif); err != nil {
		logger.Warnf(ctx, "failed to create welcome notification for user %s: %v", createdUser.Id, err)
	}

	return &createdUser, nil
}

func (s *UserService) UpdateUser(ctx context.Context, data dto.UpdateUserRequestDTO) (*domains.User, *response.Response[any]) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		)
	}

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
