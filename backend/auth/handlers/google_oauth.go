package handlers

import (
	cryptorand "crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	servererrors "sen1or/letslive/auth/errors"
	"time"
)

func (h *AuthHandler) OAuthGoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	oauthState, err := generateOAuthCookieState(w)
	if err != nil {
		h.WriteErrorResponse(w, servererrors.ErrInternalServer)
		return
	}

	u := h.googleAuthService.GenerateAuthCodeURL(oauthState)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) OAuthGoogleCallBackHandler(w http.ResponseWriter, r *http.Request) {
	GetRedirectURLOnFail := func(errMsg string) string {
		clientAddr := os.Getenv("CLIENT_URL")
		return fmt.Sprintf("%s/login?errorMessage=%s", clientAddr, errMsg)
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

	createdAuth, handleErr := h.googleAuthService.CallbackHandler(r.FormValue("code"))
	if handleErr != nil {
		http.Redirect(w, r, GetRedirectURLOnFail(handleErr.Message), http.StatusTemporaryRedirect)
		return
	}

	if err := h.setAuthJWTsInCookie(createdAuth.UserId.String(), w); err != nil {
		http.Redirect(w, r, GetRedirectURLOnFail(err.Message), http.StatusTemporaryRedirect)
		return
	}

	// redirect to main page
	http.Redirect(w, r, os.Getenv("CLIENT_URL"), http.StatusMovedPermanently)
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
