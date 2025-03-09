package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"sen1or/lets-live/auth/domains"
	servererrors "sen1or/lets-live/auth/errors"
	usergateway "sen1or/lets-live/auth/gateway/user"
	"sen1or/lets-live/auth/repositories"
	"sen1or/lets-live/pkg/logger"
	"strconv"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleAuthService struct {
	repo        repositories.AuthRepository
	userGateway usergateway.UserGateway
}

func NewGoogleAuthService(repo repositories.AuthRepository, userGateway usergateway.UserGateway) *GoogleAuthService {
	return &GoogleAuthService{
		repo:        repo,
		userGateway: userGateway,
	}
}

type googleOAuthUser struct {
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func (s GoogleAuthService) getGoogleOauthConfig() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_OAUTH_REDIRECT_URL"),
		ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func (s GoogleAuthService) GenerateAuthCodeURL(oauthState string) string {
	return s.getGoogleOauthConfig().AuthCodeURL(oauthState)
}

func (s GoogleAuthService) CallbackHandler(googleCode string) (*domains.Auth, *servererrors.ServerError) {
	data, getErr := s.getUserDataFromGoogle(googleCode)
	if getErr != nil {
		logger.Errorf("failed to get user data from google: %s", getErr)
		return nil, servererrors.ErrInternalServer
	}

	var returnedOAuthUser googleOAuthUser
	if err := json.Unmarshal(data, &returnedOAuthUser); err != nil {
		logger.Errorf("failed to unmarshal data into google user")
		return nil, servererrors.ErrInternalServer
	}

	existedRecord, err := s.repo.GetByEmail(returnedOAuthUser.Email)
	if err != nil && !errors.Is(err, servererrors.ErrAuthNotFound) {
		return nil, servererrors.ErrInternalServer
	} else if err != nil {
		dto := &usergateway.CreateUserRequestDTO{
			Username:     generateUsername(returnedOAuthUser.Email),
			Email:        returnedOAuthUser.Email,
			IsVerified:   returnedOAuthUser.VerifiedEmail,
			AuthProvider: usergateway.ProviderGoogle,
		}

		createdUser, errRes := s.userGateway.CreateNewUser(context.Background(), *dto)
		if errRes != nil {
			logger.Errorf("failed to create new user through gateway: %s", errRes.Message)
			return nil, servererrors.NewServerError(errRes.StatusCode, errRes.Message)
		}

		newOAuthUserRecord := &domains.Auth{
			Email:  returnedOAuthUser.Email,
			UserId: createdUser.Id,
		}

		// TODO: remove created user if not able to save auth
		newlyCreatedAuthRecord, err := s.repo.Create(*newOAuthUserRecord)
		if err != nil {
			return nil, err
		}

		existedRecord = newlyCreatedAuthRecord
	}

	return existedRecord, nil
}

func (s GoogleAuthService) getUserDataFromGoogle(code string) ([]byte, error) {
	token, err := s.getGoogleOauthConfig().Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("oauth exchange failed: %s", err.Error())
	}

	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("oauth failed to fetch user info: %s", err.Error())
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response from Google API: %s", response.Status)
	}

	userData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading oauth user info: %s", err.Error())
	}

	return userData, nil
}

func generateUsername(email string) string {
	base := strings.Split(email, "@")[0]
	uniqueSuffix := strconv.Itoa(rand.IntN(10000)) // Random 4-digit number
	return base + "-gguser" + uniqueSuffix
}
