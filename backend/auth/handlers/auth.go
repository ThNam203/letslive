package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/auth/dto"
	"sen1or/letslive/auth/pkg/logger"
	serviceresponse "sen1or/letslive/auth/response"
	"sen1or/letslive/auth/services"
	"sen1or/letslive/auth/utils"
)

// TODO: put verificationGateway into config
type AuthHandler struct {
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
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var userCredentials dto.LogInRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&userCredentials); err != nil {
		writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](serviceresponse.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	ip := r.Header.Get("CF-Connecting-IP")
	if err := utils.CheckCAPTCHA(userCredentials.TurnstileToken, ip); err != nil {
		writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](serviceresponse.RES_ERR_CAPTCHA_FAILED, nil, nil, nil))
		return
	}

	auth, err := h.authService.GetUserFromCredentials(ctx, userCredentials)
	if err != nil {
		writeResponse(w, ctx, err)
		return
	}

	if err := h.setAuthJWTsInCookie(ctx, auth.UserId.String(), w); err != nil {
		writeResponse(w, ctx, err)
		return
	}

	writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](
		serviceresponse.RES_SUCC_LOGIN,
		nil,
		nil,
		nil,
	))
}

func (h *AuthHandler) RequestEmailVerificationHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var requestDTO dto.SignUpRequestVerificationRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))
		return
	}

	ip := r.Header.Get("CF-Connecting-IP")
	if err := utils.CheckCAPTCHA(requestDTO.TurnstileToken, ip); err != nil {
		writeResponse(w, ctx, err)
		return
	}

	// if an auth is already existed with the email, no point to continue
	err := h.authService.CheckIfAuthExistedForEmail(ctx, requestDTO)
	if err != nil {
		writeResponse(w, ctx, err)
		return
	}

	if err := h.verificationService.CreateOTPAndSendEmailVerification(ctx, h.verificationGateway, requestDTO.Email); err != nil {
		writeResponse(w, ctx, err)
		return
	}

	writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](
		serviceresponse.RES_SUCC_SENT_VERIFICATION,
		nil,
		nil,
		nil,
	))
}

func (h *AuthHandler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	refreshTokenCookie, err := r.Cookie("REFRESH_TOKEN")

	if err != nil {
		logger.Errorf(ctx, "get refresh token from cookie failed: %s", err)
		writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		))
		return
	}

	if len(refreshTokenCookie.Value) == 0 {
		logger.Errorf(ctx, "missing refresh token")
		writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		))
		return
	}

	accessTokenInfo, refreshErr := h.jwtService.RefreshToken(ctx, refreshTokenCookie.Value)
	if refreshErr != nil {
		writeResponse(w, ctx, refreshErr)
		return
	}

	h.setAccessTokenCookie(w, accessTokenInfo.AccessToken, accessTokenInfo.AccessTokenMaxAge)
	w.WriteHeader(http.StatusNoContent)
}

// TODO: revoke refresh token
func (h *AuthHandler) LogOutHandler(w http.ResponseWriter, r *http.Request) {
	h.setAccessTokenCookie(w, "", -1)
	h.setRefreshTokenCookie(w, "", -1)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) VerifyOTPAndSignUpHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// TODO: validate
	var requestDTO dto.SignUpRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))

		return
	}

	if verifyErr := h.verificationService.Verify(ctx, requestDTO.OTPCode, requestDTO.Email); verifyErr != nil {
		writeResponse(w, ctx, verifyErr)
		return
	}

	createdAuth, err := h.authService.CreateNewAuth(ctx, requestDTO)
	if err != nil {
		writeResponse(w, ctx, err)
		return
	}

	if err := h.setAuthJWTsInCookie(ctx, createdAuth.UserId.String(), w); err != nil {
		writeResponse(w, ctx, err)
		return
	}

	writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](
		serviceresponse.RES_SUCC_SIGN_UP,
		nil,
		nil,
		nil,
	))
}

func (h *AuthHandler) UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	userUUID, err := h.getUserIDFromCookie(r)
	if err != nil {
		writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		))
		return
	}

	reqDTO := dto.ChangePasswordRequestDTO{}
	if err := json.NewDecoder(r.Body).Decode(&reqDTO); err != nil {
		writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))
		return
	}
	defer r.Body.Close()

	if err := h.authService.UpdatePassword(ctx, reqDTO, *userUUID); err != nil {
		writeResponse(w, ctx, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
