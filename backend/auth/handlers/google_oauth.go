package handlers

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"sen1or/lets-live/auth/domains"
	"sen1or/lets-live/auth/dto"
	"sen1or/lets-live/auth/repositories"
	"sen1or/lets-live/pkg/logger"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type googleOAuthUser struct {
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func getGoogleOauthConfig() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  "http://localhost:8000/auth/google/callback",
		ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func (h *AuthHandler) OAuthGoogleLogin(w http.ResponseWriter, r *http.Request) {
	oauthState, err := generateOAuthCookieState(w)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	u := getGoogleOauthConfig().AuthCodeURL(oauthState)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func generateOAuthCookieState(w http.ResponseWriter) (string, error) {
	var expiration = time.Now().Add(1 * time.Hour)

	b := make([]byte, 16)
	_, err := cryptorand.Read(b)
	if err != nil {
		return "", err
	}

	state := base64.URLEncoding.EncodeToString(b)
	cookie := &http.Cookie{Name: "oauthstate", Value: state, Expires: expiration, Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode}
	http.SetCookie(w, cookie)

	return state, nil
}

func (h *AuthHandler) OAuthGoogleCallBack(w http.ResponseWriter, r *http.Request) {
	clientAddr := os.Getenv("CLIENT_URL")
	urlDirectOnFail := clientAddr + "/auth/login"

	oauthStateCookie, err := r.Cookie("oauthstate")
	if err != nil {
		h.SetError(w, fmt.Errorf("missing OAuth state cookie"))
		http.Redirect(w, r, urlDirectOnFail, http.StatusTemporaryRedirect)
		return
	}
	oauthState := oauthStateCookie.Value

	if r.FormValue("state") != oauthState {
		h.SetError(w, fmt.Errorf("invalid state, csrf attack?"))
		http.Redirect(w, r, urlDirectOnFail, http.StatusTemporaryRedirect)
		return
	}

	data, err := getUserDataFromGoogle(r.FormValue("code"))
	if err != nil {
		h.SetError(w, fmt.Errorf("can't get user data"))
		http.Redirect(w, r, urlDirectOnFail, http.StatusTemporaryRedirect)
		return
	}

	var returnedOAuthUser googleOAuthUser
	if err := json.Unmarshal(data, &returnedOAuthUser); err != nil {
		h.SetError(w, fmt.Errorf("user data format not valid"))
		http.Redirect(w, r, urlDirectOnFail, http.StatusTemporaryRedirect)
		return
	}

	// Check if the user had already singed up before
	existedRecord, err := h.authCtrl.GetByEmail(returnedOAuthUser.Email)
	if err != nil && !errors.Is(err, repositories.ErrRecordNotFound) {
		logger.Errorf("database error: %s", err)
		h.SetError(w, fmt.Errorf("database error: %v", err))
		http.Redirect(w, r, urlDirectOnFail, http.StatusTemporaryRedirect)
		return
	} else if err != nil {
		dto := &dto.CreateUserRequestDTO{
			Username:   generateUsername(returnedOAuthUser.Email),
			Email:      returnedOAuthUser.Email,
			IsVerified: returnedOAuthUser.VerifiedEmail,
		}

		createdUser, errRes := h.userGateway.CreateNewUser(context.Background(), *dto)
		if errRes != nil {
			h.WriteErrorResponse(w, errRes.StatusCode, errors.New(errRes.Message))
			return
		}

		newOAuthUserRecord := &domains.Auth{
			Email:  returnedOAuthUser.Email,
			UserId: createdUser.Id,
		}

		// TODO: remove created user if not able to save auth
		newlyCreatedAuthRecord, err := h.authCtrl.Create(*newOAuthUserRecord)
		if err != nil {
			h.SetError(w, fmt.Errorf("error while saving auth"))
			http.Redirect(w, r, urlDirectOnFail, http.StatusTemporaryRedirect)
			return
		}

		existedRecord = newlyCreatedAuthRecord
	}

	if err := h.setAuthJWTsInCookie(existedRecord.UserId.String(), w); err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, clientAddr, http.StatusMovedPermanently)
}

func getUserDataFromGoogle(code string) ([]byte, error) {
	token, err := getGoogleOauthConfig().Exchange(context.Background(), code)
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
	return base + uniqueSuffix
}
