package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofrs/uuid/v5"
)

func (a *api) GetUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	userUUID, err := uuid.FromString(userId)

	if err != nil {
		a.errorResponse(w, http.StatusBadRequest, errors.New("userId not valid"))
	}
	user, err := a.userRepo.GetByID(userUUID)

	if err != nil {
		a.errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
