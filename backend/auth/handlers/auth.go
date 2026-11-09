package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/dto"
	serviceresponse "sen1or/letslive/auth/response"
	"sen1or/letslive/auth/services"
	"sen1or/letslive/auth/utils"
	"sen1or/letslive/shared/pkg/logger"
)

// TODO: put verificationGateway into config
type AuthHandler struct {
	jwtService          services.JWTService
	authService         services.AuthService
	googleAuthService   services.GoogleAuthService
	verificationService services.VerificationService
	verificationGateway string

	// appCheck is nil when App Check is off (FIREBASE_PROJECT_ID unset).
	appCheck     utils.AppCheckVerifier
	checkCaptcha func(token, userIPAddress string) error
}

func NewAuthHandler(
	jwtService services.JWTService,
	authService services.AuthService,
	verificationService services.VerificationService,
	googleAuthService services.GoogleAuthService,
	verficationGateway string,
	appCheck utils.AppCheckVerifier,
) *AuthHandler {
	return &AuthHandler{
		appCheck:            appCheck,
		checkCaptcha:        utils.CheckCAPTCHA,
		authService:         authService,
		googleAuthService:   googleAuthService,
		verificationService: verificationService,
		jwtService:          jwtService,
		verificationGateway: verficationGateway,
	}
}

// AppCheckHeader carries the Firebase App Check token sent by the mobile app.
const AppCheckHeader = "X-Firebase-AppCheck"

// verifyNotBot decides whether a request may proceed past the bot check.
//
//   - App Check configured and the header present: the token alone decides.
//     Valid passes (Turnstile is not consulted); invalid is rejected even if a
//     valid Turnstile token came along (fail closed, no fall-through).
//   - Otherwise (App Check off, or no header): Turnstile is required.
//
// Nothing here looks at the User-Agent, which any client can set.
func (h *AuthHandler) verifyNotBot(ctx context.Context, r *http.Request, turnstileToken string) error {
	if appCheckToken := r.Header.Get(AppCheckHeader); h.appCheck != nil && appCheckToken != "" {
		if err := h.appCheck.Verify(ctx, appCheckToken); err != nil {
			// A rejected token is a client problem, not a server fault.
			logger.Warnf(ctx, "app check verification failed: %s", err)
			return domains.ErrCaptchaFailed
		}
		return nil
	}

	return h.checkCaptcha(turnstileToken, r.Header.Get("CF-Connecting-IP"))
}

func (h *AuthHandler) LogInHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var userCredentials dto.LogInRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&userCredentials); err != nil {
		writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](serviceresponse.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	if err := h.verifyNotBot(ctx, r, userCredentials.TurnstileToken); err != nil {
		writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](serviceresponse.RES_ERR_CAPTCHA_FAILED, nil, nil, nil))
		return
	}

	auth, err := h.authService.GetUserFromCredentials(ctx, userCredentials)
	if err != nil {
		writeResponse(w, ctx, serviceresponse.FromError(err))
		return
	}

	if err := h.setAuthJWTsInCookie(ctx, auth.UserId.String(), w); err != nil {
		writeResponse(w, ctx, serviceresponse.FromError(err))
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

	if err := h.verifyNotBot(ctx, r, requestDTO.TurnstileToken); err != nil {
		writeResponse(w, ctx, serviceresponse.FromError(err))
		return
	}

	// if an auth is already existed with the email, no point to continue
	err := h.authService.CheckIfAuthExistedForEmail(ctx, requestDTO)
	if err != nil {
		writeResponse(w, ctx, serviceresponse.FromError(err))
		return
	}

	if err := h.verificationService.CreateOTPAndSendEmailVerification(ctx, h.verificationGateway, requestDTO.Email); err != nil {
		writeResponse(w, ctx, serviceresponse.FromError(err))
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
		writeResponse(w, ctx, serviceresponse.FromError(refreshErr))
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
		writeResponse(w, ctx, serviceresponse.FromError(verifyErr))
		return
	}

	createdAuth, err := h.authService.CreateNewAuth(ctx, requestDTO)
	if err != nil {
		writeResponse(w, ctx, serviceresponse.FromError(err))
		return
	}

	if err := h.setAuthJWTsInCookie(ctx, createdAuth.UserId.String(), w); err != nil {
		writeResponse(w, ctx, serviceresponse.FromError(err))
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
		writeResponse(w, ctx, serviceresponse.FromError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// JWKSHandler publishes the access-token public key. It is read by the other
// services and the gateway, never by browsers, and holds nothing secret.
func (h *AuthHandler) JWKSHandler(w http.ResponseWriter, r *http.Request) {
	body, err := h.jwtService.JWKS()
	if err != nil {
		logger.Errorf(r.Context(), "failed to render jwks: %v", err)
		writeResponse(w, r.Context(), serviceresponse.NewResponseFromTemplate[any](serviceresponse.RES_ERR_INTERNAL_SERVER, nil, nil, nil))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Write(body)
}
