package services

import (
	"context"
	"errors"
	"fmt"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/dto"
	usergateway "sen1or/letslive/auth/gateway/user"
	usergatewaydto "sen1or/letslive/auth/gateway/user/dto"
	"sen1or/letslive/auth/utils"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo          domains.AuthRepository
	signUpOTPRepo domains.SignUpOTPRepository
	userGateway   usergateway.UserGateway
}

func NewAuthService(repo domains.AuthRepository, userGateway usergateway.UserGateway) *AuthService {
	return &AuthService{
		repo:        repo,
		userGateway: userGateway,
	}
}

func (s AuthService) GetUserById(ctx context.Context, userId uuid.UUID) (*domains.Auth, error) {
	auth, err := s.repo.GetByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}

	return auth, nil
}

func (s AuthService) GetUserFromCredentials(ctx context.Context, credentials dto.LogInRequestDTO) (*domains.Auth, error) {
	validateErr := utils.Validator.Struct(&credentials)

	if validateErr != nil {
		return nil, fmt.Errorf("%w: %w", domains.ErrInvalidInput, validateErr)
	}

	auth, err := s.repo.GetByEmail(ctx, credentials.Email)
	if err != nil {
		if errors.Is(err, domains.ErrAuthNotFound) {
			return nil, domains.ErrEmailOrPasswordIncorrect
		}

		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(auth.PasswordHash), []byte(credentials.Password)); err != nil {
		return nil, domains.ErrEmailOrPasswordIncorrect
	}

	return auth, nil
}

func (s AuthService) CreateNewAuth(ctx context.Context, userForm dto.SignUpRequestDTO) (*domains.Auth, error) {
	err := utils.Validator.Struct(&userForm)
	if err != nil {
		logger.Errorf(ctx, "failed to validate user signup form data: %s", err)
		return nil, fmt.Errorf("%w: %w", domains.ErrInvalidInput, err)
	}

	existed, _ := s.repo.GetByEmail(ctx, userForm.Email)
	if existed != nil {
		return nil, domains.ErrAuthAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userForm.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf(ctx, "failed to generate hash password: %s", err)
		return nil, domains.ErrInternal
	}

	userDTO := &usergatewaydto.CreateUserRequestDTO{
		Username:     userForm.Username,
		Email:        userForm.Email,
		AuthProvider: usergatewaydto.ProviderLocal,
	}

	createdUser, errRes := s.userGateway.CreateNewUser(ctx, *userDTO)
	if errRes != nil {
		return nil, errRes
	}

	auth := &domains.Auth{
		UserId:       &createdUser.Id,
		Email:        userForm.Email,
		PasswordHash: string(hashedPassword),
	}

	createdAuthDTO, createErr := s.repo.Create(ctx, *auth)
	if createErr != nil {
		// TODO: remove user if not create auth successfully
		return nil, createErr
	}

	return createdAuthDTO, nil
}

func (s AuthService) CheckIfAuthExistedForEmail(ctx context.Context, emailVerificationForm dto.SignUpRequestVerificationRequestDTO) error {
	err := utils.Validator.Struct(&emailVerificationForm)
	if err != nil {
		logger.Errorf(ctx, "failed to validate user sign up form data: %s", err)
		return domains.ErrInvalidInput
	}

	existed, rErr := s.repo.GetByEmail(ctx, emailVerificationForm.Email)
	if rErr != nil && !errors.Is(rErr, domains.ErrAuthNotFound) {
		return rErr
	}

	if existed != nil {
		return domains.ErrAuthAlreadyExists
	}

	return nil
}

func (s AuthService) UpdatePassword(ctx context.Context, dto dto.ChangePasswordRequestDTO, userUUID uuid.UUID) error {
	if err := utils.Validator.Struct(&dto); err != nil {
		return domains.ErrInvalidInput
	}

	auth, err := s.repo.GetByUserID(ctx, userUUID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(auth.PasswordHash), []byte(dto.OldPassword)); err != nil {
		return domains.ErrPasswordNotMatch
	}

	updateHashedPassword, genErr := bcrypt.GenerateFromPassword([]byte(dto.NewPassword), bcrypt.DefaultCost)
	if genErr != nil {
		return domains.ErrInternal
	}

	auth.PasswordHash = string(updateHashedPassword)
	if err := s.repo.UpdatePasswordHash(ctx, auth.Id.String(), auth.PasswordHash); err != nil {
		return err
	}

	return nil
}

func (s AuthService) GetUserStatus(ctx context.Context, userId uuid.UUID) (string, error) {
	return s.userGateway.GetUserStatus(ctx, userId.String())
}

func (s AuthService) ReactivateUser(ctx context.Context, userId string) error {
	return s.userGateway.UpdateUserStatus(ctx, userId, usergateway.UserStatusNormal)
}
