package services

import (
	"context"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/dto"
	usergateway "sen1or/letslive/auth/gateway/user"
	"sen1or/letslive/auth/pkg/logger"
	serviceresponse "sen1or/letslive/auth/response"
	"sen1or/letslive/auth/utils"

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

func (s AuthService) GetUserById(ctx context.Context, userId uuid.UUID) (*domains.Auth, *serviceresponse.Response[any]) {
	auth, err := s.repo.GetByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}

	return auth, nil
}

func (s AuthService) GetUserFromCredentials(ctx context.Context, credentials dto.LogInRequestDTO) (*domains.Auth, *serviceresponse.Response[any]) {
	validateErr := utils.Validator.Struct(&credentials)

	if validateErr != nil {
		return nil, serviceresponse.NewResponseWithValidationErrors[any](nil, nil, validateErr)
	}

	auth, err := s.repo.GetByEmail(ctx, credentials.Email)
	if err != nil {
		if err.Code == serviceresponse.RES_ERR_AUTH_NOT_FOUND_CODE {
			return nil, serviceresponse.NewResponseFromTemplate(
				serviceresponse.RES_ERR_EMAIL_OR_PASSWORD_INCORRECT,
				err.Data,
				err.Meta,
				err.ErrorDetails,
			)
		}

		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(auth.PasswordHash), []byte(credentials.Password)); err != nil {
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_EMAIL_OR_PASSWORD_INCORRECT,
			nil,
			nil,
			nil,
		)
	}

	return auth, nil
}

func (s AuthService) CreateNewAuth(ctx context.Context, userForm dto.SignUpRequestDTO) (*domains.Auth, *serviceresponse.Response[any]) {
	err := utils.Validator.Struct(&userForm)
	if err != nil {
		logger.Errorf(ctx, "failed to validate user signup form data: %s", err)
		return nil, serviceresponse.NewResponseWithValidationErrors[any](
			nil,
			nil,
			err,
		)
	}

	existed, _ := s.repo.GetByEmail(ctx, userForm.Email)
	if existed != nil {
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_AUTH_ALREADY_EXISTS,
			nil,
			nil,
			&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"email": userForm.Email}},
		)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userForm.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf(ctx, "failed to generate hash password: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	dto := &usergateway.CreateUserRequestDTO{
		Username:     userForm.Username,
		Email:        userForm.Email,
		AuthProvider: usergateway.ProviderLocal,
	}

	createdUser, errRes := s.userGateway.CreateNewUser(ctx, *dto)
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

func (s AuthService) CheckIfAuthExistedForEmail(ctx context.Context, emailVerificationForm dto.SignUpRequestVerificationRequestDTO) *serviceresponse.Response[any] {
	err := utils.Validator.Struct(&emailVerificationForm)
	if err != nil {
		logger.Errorf(ctx, "failed to validate user sign up form data: %s", err)
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		)
	}

	existed, rErr := s.repo.GetByEmail(ctx, emailVerificationForm.Email)
	if rErr != nil && rErr.Code != serviceresponse.RES_ERR_AUTH_NOT_FOUND_CODE {
		return rErr
	}

	if existed != nil {
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_AUTH_ALREADY_EXISTS,
			nil,
			nil,
			nil,
		)
	}

	return nil
}

func (s AuthService) UpdatePassword(ctx context.Context, dto dto.ChangePasswordRequestDTO, userUUID uuid.UUID) *serviceresponse.Response[any] {
	if err := utils.Validator.Struct(&dto); err != nil {
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		)
	}

	auth, err := s.repo.GetByUserID(ctx, userUUID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(auth.PasswordHash), []byte(dto.OldPassword)); err != nil {
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_PASSWORD_NOT_MATCH,
			nil,
			nil,
			nil,
		)
	}

	updateHashedPassword, genErr := bcrypt.GenerateFromPassword([]byte(dto.NewPassword), bcrypt.DefaultCost)
	if genErr != nil {
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	auth.PasswordHash = string(updateHashedPassword)
	if err := s.repo.UpdatePasswordHash(ctx, auth.Id.String(), auth.PasswordHash); err != nil {
		return err
	}

	return nil
}
