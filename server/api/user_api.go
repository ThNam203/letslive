package api

import (
	"encoding/json"
	"net/http"
	"sen1or/lets-live/server/domain"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/bcrypt"
)

func (a *api) GetUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	user, err := a.userRepo.GetByID(userId)

	if err != nil {
		a.errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

type signUpForm struct {
	Username string `validate:"required,gte=6,lte=50"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,gte=8,lte=20"`
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
