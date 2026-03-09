package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"sen1or/letslive/auth/domains"
	usergateway "sen1or/letslive/auth/gateway/user"
	usergatewaydto "sen1or/letslive/auth/gateway/user/dto"
	"sen1or/letslive/auth/pkg/logger"
	serviceresponse "sen1or/letslive/auth/response"
	"strconv"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleAuthService struct {
	repo        domains.AuthRepository
	userGateway usergateway.UserGateway
}

func NewGoogleAuthService(repo domains.AuthRepository, userGateway usergateway.UserGateway) *GoogleAuthService {
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

func (s GoogleAuthService) CallbackHandler(ctx context.Context, googleCode string) (*domains.Auth, *serviceresponse.Response[any]) {
	data, getErr := s.getUserDataFromGoogle(ctx, googleCode)
	if getErr != nil {
		logger.Errorf(ctx, "failed to get user data from google: %s", getErr)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	var returnedOAuthUser googleOAuthUser
	if err := json.Unmarshal(data, &returnedOAuthUser); err != nil {
		logger.Errorf(ctx, "failed to unmarshal data into google user")
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	return s.findOrCreateGoogleUser(ctx, returnedOAuthUser)
}

// VerifyIDTokenAndGetUser verifies a Google ID token (from mobile's google_sign_in)
// by calling Google's tokeninfo endpoint, then finds or creates the user.
func (s GoogleAuthService) VerifyIDTokenAndGetUser(ctx context.Context, idToken string) (*domains.Auth, *serviceresponse.Response[any]) {
	tokenInfoURL := "https://oauth2.googleapis.com/tokeninfo?id_token=" + idToken
	resp, err := http.Get(tokenInfoURL)
	if err != nil {
		logger.Errorf(ctx, "failed to verify google id token: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER, nil, nil, nil,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf(ctx, "google tokeninfo returned status %d", resp.StatusCode)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_UNAUTHORIZED, nil, nil, nil,
		)
	}

	var tokenInfo struct {
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Aud           string `json:"aud"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		logger.Errorf(ctx, "failed to decode google tokeninfo: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER, nil, nil, nil,
		)
	}

	// Verify the token was issued for our client
	expectedClientID := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	if tokenInfo.Aud != expectedClientID {
		// Also check the mobile-specific client ID (iOS/Android may differ)
		mobileClientID := os.Getenv("GOOGLE_OAUTH_MOBILE_CLIENT_ID")
		if mobileClientID == "" || tokenInfo.Aud != mobileClientID {
			logger.Errorf(ctx, "google id token audience mismatch: got %s", tokenInfo.Aud)
			return nil, serviceresponse.NewResponseFromTemplate[any](
				serviceresponse.RES_ERR_UNAUTHORIZED, nil, nil, nil,
			)
		}
	}

	if tokenInfo.Email == "" || tokenInfo.EmailVerified != "true" {
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_UNAUTHORIZED, nil, nil, nil,
		)
	}

	return s.findOrCreateGoogleUser(ctx, googleOAuthUser{
		Email:         tokenInfo.Email,
		VerifiedEmail: true,
		Name:          tokenInfo.Name,
		Picture:       tokenInfo.Picture,
	})
}

// findOrCreateGoogleUser looks up a user by email; if not found, creates a new one.
func (s GoogleAuthService) findOrCreateGoogleUser(ctx context.Context, oauthUser googleOAuthUser) (*domains.Auth, *serviceresponse.Response[any]) {
	existedRecord, err := s.repo.GetByEmail(ctx, oauthUser.Email)
	if err == nil {
		return existedRecord, nil
	}

	// create new user if not found
	if err.Code == serviceresponse.RES_ERR_AUTH_NOT_FOUND_CODE {
		userDTO := &usergatewaydto.CreateUserRequestDTO{
			Username:     generateUsername(oauthUser.Email),
			Email:        oauthUser.Email,
			AuthProvider: usergatewaydto.ProviderGoogle,
		}

		createdUser, errRes := s.userGateway.CreateNewUser(ctx, *userDTO)
		if errRes != nil {
			logger.Errorf(ctx, "failed to create new user through gateway: %s", errRes.Message)
			return nil, errRes
		}

		newOAuthUserRecord := &domains.Auth{
			Email:  oauthUser.Email,
			UserId: &createdUser.Id,
		}

		// TODO: remove created user if not able to save auth
		newlyCreatedAuthRecord, err := s.repo.Create(ctx, *newOAuthUserRecord)
		if err != nil {
			return nil, err
		}

		return newlyCreatedAuthRecord, nil
	}

	return nil, serviceresponse.NewResponseFromTemplate[any](
		serviceresponse.RES_ERR_INTERNAL_SERVER,
		nil,
		nil,
		nil,
	)
}

func (s GoogleAuthService) getUserDataFromGoogle(ctx context.Context, code string) ([]byte, error) {
	token, err := s.getGoogleOauthConfig().Exchange(ctx, code)
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
	return base + "-gg" + uniqueSuffix
}
