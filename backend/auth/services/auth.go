package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/dto"
	servererrors "sen1or/letslive/auth/errors"
	usergateway "sen1or/letslive/auth/gateway/user"
	"sen1or/letslive/auth/pkg/logger"
	"sen1or/letslive/auth/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo        domains.AuthRepository
	userGateway usergateway.UserGateway
}

func NewAuthService(repo domains.AuthRepository, userGateway usergateway.UserGateway) *AuthService {
	return &AuthService{
		repo:        repo,
		userGateway: userGateway,
	}
}

func (s AuthService) GetUserById(userId uuid.UUID) (*domains.Auth, *servererrors.ServerError) {
	auth, err := s.repo.GetByUserID(userId)
	if err != nil {
		return nil, err
	}

	return auth, nil
}

func (s AuthService) GetUserFromCredentials(credentials dto.LogInRequestDTO) (*domains.Auth, *servererrors.ServerError) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validateErr := validate.Struct(&credentials)

	if validateErr != nil {
		return nil, servererrors.ErrInvalidInput
	}

	auth, err := s.repo.GetByEmail(credentials.Email)
	if err != nil {
		if errors.Is(err, servererrors.ErrAuthNotFound) {
			return nil, servererrors.ErrEmailOrPasswordIncorrect
		}

		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(auth.PasswordHash), []byte(credentials.Password)); err != nil {
		return nil, servererrors.ErrEmailOrPasswordIncorrect
	}

	return auth, nil
}

func (s AuthService) CheckCAPTCHA(token string, userIPAddress string) *servererrors.ServerError {
	formData := url.Values{}
	formData.Set("secret", os.Getenv("CLOUDFLARE_TURNSTILE_SECRET_KEY"))
	formData.Set("response", token)
	if len(userIPAddress) == 0 {
		formData.Set("remoteip", userIPAddress)
	}

	// Send verification request to Cloudflare
	resp, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", formData)
	if err != nil {
		return servererrors.ErrInternalServer
	}
	defer resp.Body.Close()

	type TurnstileResponse struct {
		Success bool `json:"success"`
	}

	// Parse response
	var outcome TurnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&outcome); err != nil {
		return servererrors.ErrInternalServer
	}

	if outcome.Success {
		return nil
	}

	return servererrors.ErrCaptchaFailed
}

func (s AuthService) CreateNewAuth(userForm dto.SignUpRequestDTO) (*domains.Auth, *servererrors.ServerError) {
	err := utils.Validator.Struct(&userForm)
	if err != nil {
		logger.Errorf("failed to validate user signup form data: %s", err)
		return nil, servererrors.ErrInvalidInput
	}

	existed, _ := s.repo.GetByEmail(userForm.Email)
	if existed != nil {
		return nil, servererrors.ErrAuthAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userForm.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("failed to generate hash password: %s", err)
		return nil, servererrors.ErrInternalServer
	}

	dto := &usergateway.CreateUserRequestDTO{
		Username:     userForm.Username,
		Email:        userForm.Email,
		IsVerified:   false,
		AuthProvider: usergateway.ProviderLocal,
	}

	createdUser, errRes := s.userGateway.CreateNewUser(context.Background(), *dto)
	if errRes != nil {
		return nil, servererrors.NewServerError(errRes.StatusCode, errRes.Message)
	}

	auth := &domains.Auth{
		UserId:       createdUser.Id,
		Email:        userForm.Email,
		PasswordHash: string(hashedPassword),
	}

	createdAuthDTO, createErr := s.repo.Create(*auth)
	if createErr != nil {
		// TODO: remove user if not create auth successfully
		return nil, createErr
	}

	return createdAuthDTO, nil
}

func (s AuthService) UpdatePassword(dto dto.ChangePasswordRequestDTO, userUUID uuid.UUID) *servererrors.ServerError {
	if err := utils.Validator.Struct(&dto); err != nil {
		return servererrors.ErrInvalidInput
	}

	auth, err := s.repo.GetByUserID(userUUID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(auth.PasswordHash), []byte(dto.OldPassword)); err != nil {
		return servererrors.ErrPasswordNotMatch
	}

	updateHashedPassword, genErr := bcrypt.GenerateFromPassword([]byte(dto.NewPassword), bcrypt.DefaultCost)
	if genErr != nil {
		return servererrors.ErrInternalServer
	}

	auth.PasswordHash = string(updateHashedPassword)
	if err := s.repo.UpdatePasswordHash(auth.Id.String(), auth.PasswordHash); err != nil {
		return err
	}

	return nil
}
