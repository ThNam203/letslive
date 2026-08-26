package handlers

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"
	"os"
	"sen1or/letslive/auth/dto"
	serviceresponse "sen1or/letslive/auth/response"
	"time"
)

func (h *AuthHandler) OAuthGoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	oauthState, err := generateOAuthCookieState(w)
	if err != nil {
		writeResponse(
			w,
			r.Context(),
			serviceresponse.NewResponseFromTemplate[any](
				serviceresponse.RES_ERR_INTERNAL_SERVER,
				nil,
				nil,
				nil,
			),
		)
		return
	}

	u := h.googleAuthService.GenerateAuthCodeURL(oauthState)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) OAuthGoogleCallBackHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	GetRedirectURLOnFail := func(errMsg string) string {
		clientAddr := os.Getenv("CLIENT_URL")
		return fmt.Sprintf("%s/login?errorMessage=%s", clientAddr, errMsg)
	}

	GetRedirectURLOnSuccess := func(redirectUrl string) string {
		clientAddr := os.Getenv("CLIENT_URL")
		return fmt.Sprintf("%s/login?redirectUrl=%s", clientAddr, redirectUrl)
	}

	oauthStateCookie, err := r.Cookie("oauthstate")
	if err != nil {
		http.Redirect(w, r, GetRedirectURLOnFail("Missing OAuth state cookie"), http.StatusTemporaryRedirect)
		return
	}
	oauthState := oauthStateCookie.Value

	if r.FormValue("state") != oauthState {
		http.Redirect(w, r, GetRedirectURLOnFail("Invalid state"), http.StatusTemporaryRedirect)
		return
	}

	createdAuth, handleErr := h.googleAuthService.CallbackHandler(ctx, r.FormValue("code"))
	if handleErr != nil {
		http.Redirect(w, r, GetRedirectURLOnFail(handleErr.Message), http.StatusTemporaryRedirect)
		return
	}

	isDisabled, reactivationToken, statusErr := h.checkAccountStatus(ctx, *createdAuth.UserId)
	if statusErr != nil {
		http.Redirect(w, r, GetRedirectURLOnFail(statusErr.Message), http.StatusTemporaryRedirect)
		return
	}

	if isDisabled {
		http.Redirect(w, r, buildDisabledRedirectURL(os.Getenv("CLIENT_URL"), reactivationToken), http.StatusTemporaryRedirect)
		return
	}

	if err := h.setAuthJWTsInCookie(ctx, createdAuth.UserId.String(), w); err != nil {
		http.Redirect(w, r, GetRedirectURLOnFail(err.Message), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, GetRedirectURLOnSuccess("/account-setup"), http.StatusMovedPermanently)
}

func buildDisabledRedirectURL(clientAddr, reactivationToken string) string {
	return fmt.Sprintf("%s/login?accountDisabled=true&reactivationToken=%s", clientAddr, neturl.QueryEscape(reactivationToken))
}

// OAuthGoogleMobileHandler handles Google sign-in from mobile clients.
// Mobile sends a Google ID token obtained via the native google_sign_in SDK.
func (h *AuthHandler) OAuthGoogleMobileHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var body struct {
		IDToken string `json:"idToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.IDToken == "" {
		writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INVALID_PAYLOAD, nil, nil, nil,
		))
		return
	}

	createdAuth, authErr := h.googleAuthService.VerifyIDTokenAndGetUser(ctx, body.IDToken)
	if authErr != nil {
		writeResponse(w, ctx, authErr)
		return
	}

	isDisabled, reactivationToken, statusErr := h.checkAccountStatus(ctx, *createdAuth.UserId)
	if statusErr != nil {
		writeResponse(w, ctx, statusErr)
		return
	}

	if isDisabled {
		disabledData := any(dto.AccountDisabledResponseDTO{ReactivationToken: reactivationToken})
		writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate(serviceresponse.RES_ERR_ACCOUNT_DISABLED, &disabledData, nil, nil))
		return
	}

	if err := h.setAuthJWTsInCookie(ctx, createdAuth.UserId.String(), w); err != nil {
		writeResponse(w, ctx, err)
		return
	}

	writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](
		serviceresponse.RES_SUCC_LOGIN, nil, nil, nil,
	))
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
