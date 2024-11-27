package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"sen1or/lets-live/auth/controllers"
	"sen1or/lets-live/auth/domains"
	"sen1or/lets-live/auth/dto"
	usergateway "sen1or/lets-live/auth/gateway/user/http"
	"sen1or/lets-live/auth/utils"
	userdto "sen1or/lets-live/user/dto"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandlerConfig struct {
	RefreshTokenExpiresDuration string
	AccessTokenExpiresDuration  string
}

type AuthHandler struct {
	ErrorHandler
	refreshTokenCtrl *controllers.RefreshTokenController
	authCtrl         *controllers.AuthController
	verifyTokenCtrl  *controllers.VerifyTokenController
	gateway          *usergateway.UserGateway
	authServerURL    string
	config           AuthHandlerConfig
}

func NewAuthHandler(
	refreshTokenCtrl *controllers.RefreshTokenController,
	authCtrl *controllers.AuthController,
	verifyTokenCtrl *controllers.VerifyTokenController,
	authServerURL string,
	config AuthHandlerConfig,
	gateway *usergateway.UserGateway,
) *AuthHandler {
	return &AuthHandler{
		authCtrl:         authCtrl,
		verifyTokenCtrl:  verifyTokenCtrl,
		refreshTokenCtrl: refreshTokenCtrl,
		authServerURL:    authServerURL,
		config:           config,
		gateway:          gateway,
	}
}

// LogInHandler handles user login.
// @Summary Log in a user
// @Description Authenticate user with email and password
// @Tags Authentication
// @Accept  json
// @Param   userCredentials  body  logInForm  true  "User credentials"
// @Success 204
// @Header 204 {string} refreshToken "Refresh Token"
// @Header 204 {string} accessToken "Access Token"
// @Failure 400 {string} string "Invalid body"
// @Failure 401 {string} string "Username or password is not correct"
// @Router /auth/login [post]
func (h *AuthHandler) LogInHandler(w http.ResponseWriter, r *http.Request) {
	var userCredentials dto.LogInRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&userCredentials); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(&userCredentials)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	user, err := h.authCtrl.GetByEmail(userCredentials.Email)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(userCredentials.Password)); err != nil {
		h.WriteErrorResponse(w, http.StatusUnauthorized, errors.New("username or password is not correct!"))
		return
	}

	if err := h.setAuthJWTsInCookie(user.ID, w); err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SignUpHandler handles user registration.
// @Summary Sign up a new user
// @Description Register a new user with username, email, and password
// @Description On success, redirect user to index page and set refresh and access token in cookie
// @Tags Authentication
// @Accept  json
// @Param userForm body signUpForm true "User registration data"
// @Success 204
// @Header 204 {string} refreshToken "Refresh Token"
// @Header 204 {string} accessToken "Access Token"
// @Failure 400 {object} HTTPErrorResponse
// @Failure 500 {object} HTTPErrorResponse
// @Router /auth/signup [post]
func (h *AuthHandler) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var userForm dto.SignUpRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&userForm); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	err := utils.Validator.Struct(&userForm)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userForm.Password), bcrypt.DefaultCost)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	dto := &userdto.CreateUserRequestDTO{
		Username: userForm.Username,
		Email:    userForm.Email,
	}

	createdUser, err := h.gateway.CreateNewUser(context.Background(), *dto)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	auth := &domains.Auth{
		UserID:       createdUser.ID,
		Email:        userForm.Email,
		PasswordHash: string(hashedPassword),
		IsVerified:   false,
	}

	createdAuthDTO, err := h.authCtrl.Create(*auth)
	createdAuthDTO.Username = createdUser.Username

	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	go h.sendConfirmEmail(auth.UserID, auth.Email, h.authServerURL)

	if err := h.setAuthJWTsInCookie(auth.UserID, w); err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdAuthDTO)
}

func (h *AuthHandler) setAuthJWTsInCookie(userId uuid.UUID, w http.ResponseWriter) error {
	tokensInfo, err := h.refreshTokenCtrl.GenerateTokenPair(userId)
	if err != nil {
		return err
	}

	h.setTokens(w, tokensInfo.RefreshToken, tokensInfo.AccessToken, tokensInfo.RefreshTokenExpiresAt, tokensInfo.AccessTokenExpiresAt)
	return nil
}

// verifyEmailHandler verifies a user's email address.
// @Summary Verify user email
// @Description Verifies a user's email address with the provided token
// @Tags Authentication
// @Accept  json
// @Param   token  query  string  true  "Email verification token"
// @Success 200 {string} string "Return a Email verification complete! string"
// @Failure 400 {string} string "Verify token expired or invalid."
// @Failure 500 {string} string "An error occurred while verifying the user."
// @Router /auth/verify [get]
func (h *AuthHandler) VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	var token = r.URL.Query().Get("token")

	verifyToken, err := h.verifyTokenCtrl.GetByValue(token)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	user, err := h.authCtrl.GetByID(verifyToken.UserID)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	} else if user.IsVerified {
		w.Write([]byte("Your email has already been verified!"))
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if verifyToken.ExpiresAt.Before(time.Now()) {
		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("verify token expired: %s", verifyToken.ExpiresAt.Local().String()))
		return
	}

	updateVerifiedAuth := &domains.Auth{
		ID:         user.ID,
		IsVerified: true,
	}

	_, err = h.authCtrl.UpdateUserVerify(*updateVerifiedAuth)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.verifyTokenCtrl.DeleteByID(verifyToken.ID)
	w.Write([]byte("Email verification complete!"))
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) sendConfirmEmail(userId uuid.UUID, userEmail string, authServerURL string) {
	verifyToken, err := h.verifyTokenCtrl.Create(userId)
	if err != nil {
		log.Printf("error while trying to create verify token: %s", err.Error())
		return
	}

	smtpServer := "smtp.gmail.com:587"
	smtpUser := "letsliveglobal@gmail.com"
	smtpPassword := os.Getenv("GMAIL_APP_PASSWORD")

	from := "letsliveglobal@gmail.com"
	to := []string{userEmail}
	subject := "Lets Live Email Confirmation"
	body := `<!DOCTYPE html>
<html>
<head>
    <title>` + subject + `</title>
</head>
<body>
    <p>This is a test email with a clickable link.</p>
	<p>Click <a href="` + authServerURL + `/v1/auth/email-verify?token=` + verifyToken.Token + `">here</a> to confirm your email.</p>
</body>
</html>`

	msg := "From: " + from + "\n" +
		"To: " + userEmail + "\n" +
		"Subject: " + subject + "\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\n\n" +
		body

	auth := smtp.PlainAuth("", smtpUser, smtpPassword, "smtp.gmail.com")

	err = smtp.SendMail(smtpServer, auth, from, to, []byte(msg))
	if err != nil {
		log.Printf("failed trying to send confirmation email: %s", err.Error())
	}
}

func (h *AuthHandler) setTokens(w http.ResponseWriter, refreshToken string, accessToken string, refreshTokenExpiresAt time.Time, accessTokenExpiresAt time.Time) {
	refreshTokenMaxAge, _ := time.ParseDuration(h.config.RefreshTokenExpiresDuration)
	accessTokenMaxAge, _ := time.ParseDuration(h.config.AccessTokenExpiresDuration)

	http.SetCookie(w, &http.Cookie{
		Name:  "refreshToken",
		Value: refreshToken,

		Expires:  refreshTokenExpiresAt,
		MaxAge:   int(refreshTokenMaxAge.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:  "accessToken",
		Value: accessToken,

		Expires:  accessTokenExpiresAt,
		MaxAge:   int(accessTokenMaxAge.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	})
}

func (h *AuthHandler) GetAuthByUserIDHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	if len(userId) == 0 {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("missing user id"))
		return
	}

	userUUID, err := uuid.FromString(userId)

	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("userId not valid"))
	}
	user, err := h.authCtrl.GetByID(userUUID)

	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// TODO: change to update password route
//func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
//	userID := r.PathValue("id")
//	if len(userID) == 0 {
//		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("missing user id"))
//		return
//	}
//	userUUID, err := uuid.FromString(userID)
//
//	if err != nil {
//		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("userId not valid"))
//	}
//	user, err := h.ctrl.GetByID(userUUID)
//
//	if err != nil {
//		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
//		return
//	}
//
//	var requestBody domains.Auth
//	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
//		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("error decoding request body: %s", err.Error()))
//		return
//	}
//
//	requestBody.ID = userUUID
//
//	if err := h.ctrl.Update(&requestBody); err != nil {
//		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
//		return
//	}
//
//	w.Header().Set("Content-Type", "application/json")
//	w.WriteHeader(http.StatusOK)
//	json.NewEncoder(w).Encode(user)
//}
