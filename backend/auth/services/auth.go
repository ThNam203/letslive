package services

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/dto"
	usergateway "sen1or/letslive/auth/gateway/user"
	"sen1or/letslive/auth/pkg/logger"
	serviceresponse "sen1or/letslive/auth/responses"
	"sen1or/letslive/auth/utils"

	"github.com/go-playground/validator/v10"
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

func (s AuthService) GetUserById(ctx context.Context, userId uuid.UUID) (*domains.Auth, *serviceresponse.ServiceErrorResponse) {
	auth, err := s.repo.GetByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}

	return auth, nil
}

func (s AuthService) GetUserFromCredentials(ctx context.Context, credentials dto.LogInRequestDTO) (*domains.Auth, *serviceresponse.ServiceErrorResponse) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validateErr := validate.Struct(&credentials)

	if validateErr != nil {
		return nil, serviceresponse.ErrInvalidInput
	}

	auth, err := s.repo.GetByEmail(ctx, credentials.Email)
	if err != nil {
		if errors.Is(err, serviceresponse.ErrAuthNotFound) {
			return nil, serviceresponse.ErrEmailOrPasswordIncorrect
		}

		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(auth.PasswordHash), []byte(credentials.Password)); err != nil {
		return nil, serviceresponse.ErrEmailOrPasswordIncorrect
	}

	return auth, nil
}

func (s AuthService) CreateNewAuth(ctx context.Context, userForm dto.SignUpRequestDTO) (*domains.Auth, *serviceresponse.ServiceErrorResponse) {
	err := utils.Validator.Struct(&userForm)
	if err != nil {
		logger.Errorf("failed to validate user signup form data: %s", err)
		return nil, serviceresponse.ErrInvalidInput
	}

	existed, _ := s.repo.GetByEmail(ctx, userForm.Email)
	if existed != nil {
		return nil, serviceresponse.ErrAuthAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userForm.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("failed to generate hash password: %s", err)
		return nil, serviceresponse.ErrInternalServer
	}

	dto := &usergateway.CreateUserRequestDTO{
		Username:     userForm.Username,
		Email:        userForm.Email,
		AuthProvider: usergateway.ProviderLocal,
	}

	createdUser, errRes := s.userGateway.CreateNewUser(ctx, *dto)
	if errRes != nil {
		return nil, serviceresponse.NewServiceErrorResponse(errRes.StatusCode, errRes.Message)
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

func (s AuthService) CheckIfAuthExistedForEmail(ctx context.Context, emailVerificationForm dto.SignUpRequestVerificationRequestDTO) *serviceresponse.ServiceErrorResponse {
	err := utils.Validator.Struct(&emailVerificationForm)
	if err != nil {
		logger.Errorf("failed to validate user sign up form data: %s", err)
		return serviceresponse.ErrInvalidInput
	}

	existed, rErr := s.repo.GetByEmail(ctx, emailVerificationForm.Email)
	if rErr != nil && !errors.Is(rErr, serviceresponse.ErrAuthNotFound) {
		return rErr
	}

	if existed != nil {
		return serviceresponse.ErrAuthAlreadyExists
	}

	return nil
}

func (s AuthService) UpdatePassword(ctx context.Context, dto dto.ChangePasswordRequestDTO, userUUID uuid.UUID) *serviceresponse.ServiceErrorResponse {
	if err := utils.Validator.Struct(&dto); err != nil {
		return serviceresponse.ErrInvalidInput
	}

	auth, err := s.repo.GetByUserID(ctx, userUUID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(auth.PasswordHash), []byte(dto.OldPassword)); err != nil {
		return serviceresponse.ErrPasswordNotMatch
	}

	updateHashedPassword, genErr := bcrypt.GenerateFromPassword([]byte(dto.NewPassword), bcrypt.DefaultCost)
	if genErr != nil {
		return serviceresponse.ErrInternalServer
	}

	auth.PasswordHash = string(updateHashedPassword)
	if err := s.repo.UpdatePasswordHash(ctx, auth.Id.String(), auth.PasswordHash); err != nil {
		return err
	}

	return nil
}
