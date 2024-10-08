package api

import (
	"errors"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/mux"
)

func (a *api) SetUserStreamOnline(w http.ResponseWriter, r *http.Request) {
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

	w.WriteHeader(http.StatusNoContent)
}

func (a *api) SetUserStreamOffline(w http.ResponseWriter, r *http.Request) {
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
