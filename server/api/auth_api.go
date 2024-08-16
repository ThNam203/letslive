package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"sen1or/lets-live/server/domain"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/bcrypt"
)

type userSignUpForm struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,gte=8,lte=72"`
}

type signUpForm struct {
	Username string `validate:"required,gte=6,lte=50"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,gte=8,lte=72"`
}

func (a *api) LogInHandler(w http.ResponseWriter, r *http.Request) {
	var userCredentials userSignUpForm
	if err := json.NewDecoder(r.Body).Decode(&userCredentials); err != nil {
		a.errorResponse(w, http.StatusBadRequest, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(&userCredentials)
	if err != nil {
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

	a.setTokens(w, refreshToken, accessToken)

	w.WriteHeader(http.StatusOK)
}

func (a *api) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var userForm signUpForm
	json.NewDecoder(r.Body).Decode(&userForm)

	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(&userForm)
	if err != nil {
		a.errorResponse(w, http.StatusBadRequest, err)
		return
	}

	uuid, err := uuid.NewGen().NewV4()
	if err != nil {
		a.errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userForm.Password), bcrypt.DefaultCost)

	if err != nil {
		a.errorResponse(w, http.StatusInternalServerError, err)
	}

	user := &domain.User{
		ID:           uuid,
		Username:     userForm.Username,
		Email:        userForm.Email,
		PasswordHash: string(hashedPassword),
	}

	a.userRepo.Create(*user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(*user)
}
