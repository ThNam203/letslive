package handlers

import (
	"encoding/json"
	"net/http"
	"sen1or/lets-live/auth/dto"
	servererrors "sen1or/lets-live/auth/errors"
	"sen1or/lets-live/auth/services"
	"sen1or/lets-live/pkg/logger"
)

// TODO: put verificationGateway into config
type AuthHandler struct {
	ErrorHandler
	jwtService          services.JWTService
	authService         services.AuthService
	googleAuthService   services.GoogleAuthService
	verificationService services.VerificationService
	verificationGateway string
}

func NewAuthHandler(
	jwtService services.JWTService,
	authService services.AuthService,
	verificationService services.VerificationService,
	googleAuthService services.GoogleAuthService,
	verficationGateway string,
) *AuthHandler {
	return &AuthHandler{
		authService:         authService,
		googleAuthService:   googleAuthService,
		verificationService: verificationService,
		jwtService:          jwtService,
		verificationGateway: verficationGateway,
	}
}

func (h *AuthHandler) LogInHandler(w http.ResponseWriter, r *http.Request) {
	var userCredentials dto.LogInRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&userCredentials); err != nil {
		h.WriteErrorResponse(w, servererrors.ErrInvalidPayload)
		return
	}

	auth, err := h.authService.GetUserFromCredentials(userCredentials)
	if err != nil {
		h.WriteErrorResponse(w, err)
		return
	}

	if err := h.setAuthJWTsInCookie(auth.UserId.String(), w); err != nil {
		h.WriteErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var userForm dto.SignUpRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&userForm); err != nil {
		h.WriteErrorResponse(w, servererrors.ErrInvalidPayload)
		return
	}

	createdAuth, err := h.authService.CreateNewAuth(userForm)
	if err != nil {
		h.WriteErrorResponse(w, err)
		return
	}

	if err := h.setAuthJWTsInCookie(createdAuth.UserId.String(), w); err != nil {
		h.WriteErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *AuthHandler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	refreshTokenCookie, err := r.Cookie("REFRESH_TOKEN")
	if err != nil {
		logger.Errorf("get refresh token from cookie failed: %s", err)
		h.WriteErrorResponse(w, servererrors.ErrUnauthorized)
		return
	}

	if len(refreshTokenCookie.Value) == 0 {
		logger.Errorf("missing refresh token")
		h.WriteErrorResponse(w, servererrors.ErrUnauthorized)
		return
	}

	accessTokenInfo, refreshErr := h.jwtService.RefreshToken(refreshTokenCookie.Value)
	if refreshErr != nil {
		h.WriteErrorResponse(w, refreshErr)
		return
	}

	h.setAccessTokenCookie(w, accessTokenInfo.AccessToken, accessTokenInfo.AccessTokenMaxAge)
	w.WriteHeader(http.StatusNoContent)
}

// TODO: revoke refresh token
func (h *AuthHandler) LogOutHandler(w http.ResponseWriter, r *http.Request) {
	h.setAccessTokenCookie(w, "", 0)
	h.setRefreshTokenCookie(w, "", 0)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	var token = r.URL.Query().Get("token")
	if len(token) == 0 {
		w.Write([]byte("Missing token!"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if verifyErr := h.verificationService.Verify(token); verifyErr != nil {
		w.Write([]byte(verifyErr.Message))
		w.WriteHeader(verifyErr.StatusCode)
		return
	}

	w.Write([]byte("Email verification complete!"))
	w.WriteHeader(http.StatusOK)
}

// TODO: check if has been verified ?
func (h *AuthHandler) SendVerificationHandler(w http.ResponseWriter, r *http.Request) {
	userUUID, cookieErr := h.getUserIDFromCookie(r)
	if cookieErr != nil {
		h.WriteErrorResponse(w, servererrors.ErrUnauthorized)
		return
	}

	auth, err := h.authService.GetUserById(*userUUID)
	if err != nil {
		h.WriteErrorResponse(w, err)
		return
	}

	createdToken, err := h.verificationService.Create(auth.UserId)
	if err != nil {
		h.WriteErrorResponse(w, err)
		return
	}

	h.verificationService.SendConfirmEmail(*createdToken, h.verificationGateway, auth.Email)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserIDFromCookie(r)
	if err != nil {
		h.WriteErrorResponse(w, servererrors.ErrUnauthorized)
		return
	}

	reqDTO := dto.ChangePasswordRequestDTO{}
	if err := json.NewDecoder(r.Body).Decode(&reqDTO); err != nil {
		h.WriteErrorResponse(w, servererrors.ErrInvalidPayload)
		return
	}
	defer r.Body.Close()

	if err := h.authService.UpdatePassword(reqDTO, *userUUID); err != nil {
		h.WriteErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
