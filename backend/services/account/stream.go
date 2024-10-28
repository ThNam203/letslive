package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/mux"
)

func (a *APIServer) SetUserStreamOnline(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	apiKey, ok := params["apiKey"]
	if !ok {
		a.errorResponse(w, http.StatusBadRequest, errors.New("missing api key"))
		return
	}

	apiKeyUUID, err := uuid.FromString(apiKey)
	if err != nil {
		a.errorResponse(w, http.StatusBadRequest, errors.New("api key not valid"))
		return
	}

	user, err := a.userRepo.GetByAPIKey(apiKeyUUID)

	if err != nil {
		a.errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	// ignore this error
	if user.IsOnline {
		a.setError(w, errors.New("user already streaming"))
	}

	user.IsOnline = true

	if err := a.userRepo.Update(*user); err != nil {
		a.errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := struct {
		UserID string `json:"userID"`
	}{
		UserID: user.ID.String(),
	}

	w.Header().Set("Content-Type", "application/text")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (a *APIServer) SetUserStreamOffline(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	apiKey, ok := params["apiKey"]
	if !ok {
		a.errorResponse(w, http.StatusBadRequest, errors.New("missing api key"))
		return
	}

	apiKeyUUID, err := uuid.FromString(apiKey)
	if err != nil {
		a.errorResponse(w, http.StatusBadRequest, errors.New("api key not valid"))
		return
	}

	user, err := a.userRepo.GetByAPIKey(apiKeyUUID)

	if err != nil {
		a.errorResponse(w, http.StatusNotFound, errors.New("user not found with api key "+apiKey))
		return
	}

	// ignore this error
	if !user.IsOnline {
		a.setError(w, errors.New("user is not streaming"))
	}

	user.IsOnline = false

	if err := a.userRepo.Save(*user); err != nil {
		a.errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *APIServer) GetOnlineStreams(w http.ResponseWriter, r *http.Request) {
	streamingUsers, err := a.userRepo.GetStreamingUsers()
	if err != nil {
		a.errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(streamingUsers); err != nil {
		a.errorResponse(w, http.StatusInternalServerError, err)
		return
	}
}
