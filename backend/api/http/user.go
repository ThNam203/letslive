package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sen1or/lets-live/api/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/mux"
)

func (a *APIServer) GetUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userId, ok := params["id"]
	if !ok {
		a.errorResponse(w, http.StatusBadRequest, errors.New("missing user id"))
		return
	}

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

func (a *APIServer) PatchUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID, ok := params["id"]
	if !ok {
		a.errorResponse(w, http.StatusBadRequest, errors.New("missing user id"))
		return
	}
	userUUID, err := uuid.FromString(userID)

	if err != nil {
		a.errorResponse(w, http.StatusBadRequest, errors.New("userId not valid"))
	}
	user, err := a.userRepo.GetByID(userUUID)

	if err != nil {
		a.errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	var requestBody *domains.User
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		a.errorResponse(w, http.StatusBadRequest, fmt.Errorf("error decoding request body: %s", err.Error()))
		return
	}

	requestBody.ID = userUUID

	if err := a.userRepo.Update(*requestBody); err != nil {
		a.errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
