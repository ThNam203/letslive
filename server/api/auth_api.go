package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"sen1or/lets-live/server/config"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type userCredentials struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,gte=8,lte=72"`
}

func (a *api) LogInHandler(w http.ResponseWriter, r *http.Request) {
	var userCredentials userCredentials
	if err := json.NewDecoder(r.Body).Decode(&userCredentials); err != nil {
		a.errorResponse(w, http.StatusBadRequest, err)
		return
	}

	user, err := a.userRepo.GetByEmail(userCredentials.Email)
	if err != nil {
		a.errorResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(userCredentials.Password)); err != nil {
		a.errorResponse(w, http.StatusUnauthorized, errors.New("username or password is not correct!"))
		return
	}

	// TODO: sync the expires date of refresh token in database with client
	refreshToken, accessToken, err := a.refreshTokenRepo.GenerateTokenPair(user.ID)
	if err != nil {
		a.errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "refreshToken",
		Value: refreshToken,

		Expires:  time.Now().Add(config.RefreshTokenExpiresDuration),
		MaxAge:   config.RefreshTokenMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&struct {
		accessToken string
	}{
		accessToken: accessToken,
	})
}
