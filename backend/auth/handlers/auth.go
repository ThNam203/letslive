package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"sen1or/lets-live/auth/controllers"
	"sen1or/lets-live/auth/domains"
	"sen1or/lets-live/auth/dto"
	usergateway "sen1or/lets-live/auth/gateway/user"
	"sen1or/lets-live/auth/repositories"
	"sen1or/lets-live/auth/types"
	"sen1or/lets-live/auth/utils"
	"sen1or/lets-live/pkg/logger"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	ErrorHandler
	tokenCtrl           controllers.TokenController
	authCtrl            controllers.AuthController
	verifyTokenCtrl     controllers.VerifyTokenController
	userGateway         usergateway.UserGateway
	verificationGateway string
}

func NewAuthHandler(
	tokenCtrl controllers.TokenController,
	authCtrl controllers.AuthController,
	verifyTokenCtrl controllers.VerifyTokenController,
	verficationGateway string,
	userGateway usergateway.UserGateway,
) *AuthHandler {
	return &AuthHandler{
		authCtrl:            authCtrl,
		verifyTokenCtrl:     verifyTokenCtrl,
		tokenCtrl:           tokenCtrl,
		verificationGateway: verficationGateway,
		userGateway:         userGateway,
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

	auth, err := h.authCtrl.GetByEmail(userCredentials.Email)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(auth.PasswordHash), []byte(userCredentials.Password)); err != nil {
		h.WriteErrorResponse(w, http.StatusUnauthorized, errors.New("username or password is not correct!"))
		return
	}

	if err := h.setAuthJWTsInCookie(auth.UserId.String(), w); err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	refreshTokenCookie, err := r.Cookie("REFRESH_TOKEN")
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("refresh token cookie fails: %s", err))
		return
	}

	if len(refreshTokenCookie.Value) == 0 {
		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("missing refresh token"))
		return
	}

	accessTokenInfo, err := h.tokenCtrl.RefreshToken(refreshTokenCookie.Value)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("can't refresh token: %s", err.Error()))
		return
	}

	h.setAccessTokenCookie(w, accessTokenInfo.AccessToken, accessTokenInfo.AccessTokenMaxAge)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) LogOutHandler(w http.ResponseWriter, r *http.Request) {
	h.setAccessTokenCookie(w, "", 0)
	h.setRefreshTokenCookie(w, "", 0)
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

	dto := &dto.CreateUserRequestDTO{
		Username:   userForm.Username,
		Email:      userForm.Email,
		IsVerified: false,
	}

	createdUser, errRes := h.userGateway.CreateNewUser(context.Background(), *dto)
	if errRes != nil {
		h.WriteErrorResponse(w, errRes.StatusCode, errors.New(errRes.Message))
		return
	}

	auth := &domains.Auth{
		UserId:       createdUser.Id,
		Email:        userForm.Email,
		PasswordHash: string(hashedPassword),
	}

	createdAuthDTO, err := h.authCtrl.Create(*auth)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.setAuthJWTsInCookie(auth.UserId.String(), w); err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdAuthDTO)
}

func (h *AuthHandler) setAuthJWTsInCookie(userId string, w http.ResponseWriter) error {
	tokensInfo, err := h.tokenCtrl.GenerateTokenPair(userId)
	if err != nil {
		return err
	}

	h.setAccessTokenCookie(w, tokensInfo.AccessToken, tokensInfo.AccessTokenMaxAge)
	h.setRefreshTokenCookie(w, tokensInfo.RefreshToken, tokensInfo.RefreshTokenMaxAge)

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

	if verifyToken.ExpiresAt.Before(time.Now()) {
		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("verify token expired: %s", verifyToken.ExpiresAt.Local().String()))
		return
	}

	errRes := h.userGateway.UpdateUserVerified(context.Background(), verifyToken.UserID.String())
	if errRes != nil {
		h.WriteErrorResponse(w, errRes.StatusCode, errors.New(errRes.Message))
		return
	}

	h.verifyTokenCtrl.DeleteByID(verifyToken.ID)
	w.Write([]byte("Email verification complete!"))
	w.WriteHeader(http.StatusOK)
}

// TODO: check if has been verified ?
func (h *AuthHandler) SendVerificationHandler(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserIDFromCookie(r)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusUnauthorized, err)
		return
	}

	auth, err := h.authCtrl.GetByUserID(*userUUID)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.sendConfirmEmail(auth.UserId, auth.Email)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) sendConfirmEmail(userId uuid.UUID, userEmail string) {
	verifyToken, err := h.verifyTokenCtrl.Create(userId)
	if err != nil {
		logger.Errorf("error while trying to create verify token: %s", err.Error())
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
    <p>Please confirm your email address, if you did not request any verification from Let's Live, please let us know.</p>
	<p>Click <a href="` + h.verificationGateway + `/auth/email-verify?token=` + verifyToken.Token + `">here</a> to confirm your email.</p>

	<p>Best Regards</p>
	<p>Let's Live Global</p>
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
		logger.Errorf("failed trying to send confirmation email: %s", err.Error())
	}
}

func (h *AuthHandler) GetAuthByUserIDHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	if len(userId) == 0 {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("missing user id"))
		return
	}

	userUUID, err := uuid.FromString(userId)

	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("user id not valid"))
		return
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

func (h *AuthHandler) UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserIDFromCookie(r)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusUnauthorized, err)
		return
	}

	reqDTO := dto.ChangePasswordRequestDTO{}
	if err := json.NewDecoder(r.Body).Decode(&reqDTO); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	if err := utils.Validator.Struct(&reqDTO); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	auth, err := h.authCtrl.GetByUserID(*userUUID)
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			h.WriteErrorResponse(w, http.StatusNotFound, err)
		} else {
			h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(auth.PasswordHash), []byte(reqDTO.OldPassword)); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("old password does not match"))
		return
	}

	updateHashedPassword, err := bcrypt.GenerateFromPassword([]byte(reqDTO.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	// update the password
	auth.PasswordHash = string(updateHashedPassword)
	if _, err := h.authCtrl.UpdatePasswordHash(*auth); err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) setRefreshTokenCookie(w http.ResponseWriter, refreshToken string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:  "REFRESH_TOKEN",
		Value: refreshToken,

		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) setAccessTokenCookie(w http.ResponseWriter, accessToken string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:  "ACCESS_TOKEN",
		Value: accessToken,

		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

}

func (h *AuthHandler) getUserIDFromCookie(r *http.Request) (*uuid.UUID, error) {
	accessTokenCookie, err := r.Cookie("ACCESS_TOKEN")
	if err != nil || len(accessTokenCookie.Value) == 0 {
		return nil, errors.New("missing credentials")
	}

	myClaims := types.MyClaims{}
	_, _, err = jwt.NewParser().ParseUnverified(accessTokenCookie.Value, &myClaims)

	userUUID, err := uuid.FromString(myClaims.UserId)
	if err != nil {
		return nil, errors.New("user id not valid")
	}

	return &userUUID, nil
}
