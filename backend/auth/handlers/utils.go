package handlers

import (
	"errors"
	"net/http"
	servererrors "sen1or/letslive/auth/errors"
	"sen1or/letslive/auth/types"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

func (h *AuthHandler) setAuthJWTsInCookie(userId string, w http.ResponseWriter) *servererrors.ServerError {
	tokensInfo, err := h.jwtService.GenerateTokenPair(userId)
	if err != nil {
		return err
	}

	h.setAccessTokenCookie(w, tokensInfo.AccessToken, tokensInfo.AccessTokenMaxAge)
	h.setRefreshTokenCookie(w, tokensInfo.RefreshToken, tokensInfo.RefreshTokenMaxAge)

	return nil
}

func (h *AuthHandler) setRefreshTokenCookie(w http.ResponseWriter, refreshToken string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:  "REFRESH_TOKEN",
		Value: refreshToken,

		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode, // use none for cross site cookie, im using different domains for fe and be
	})
}

func (h *AuthHandler) setAccessTokenCookie(w http.ResponseWriter, accessToken string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:  "ACCESS_TOKEN",
		Value: accessToken,

		Path:        "/",
		MaxAge:      maxAge,
		HttpOnly:    true,
		Secure:      true,
		SameSite:    http.SameSiteNoneMode,
		Partitioned: true,
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
