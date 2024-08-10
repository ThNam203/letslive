package api

import (
	"encoding/json"
	"net/http"
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
