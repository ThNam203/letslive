package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sen1or/lets-live/livestream/controllers"
	"sen1or/lets-live/livestream/dto"
	"sen1or/lets-live/livestream/repositories"
	"sen1or/lets-live/livestream/utils"
	"time"

	"github.com/gofrs/uuid/v5"
)

type LivestreamHandler struct {
	ErrorHandler
	ctrl controllers.LivestreamController
}

func NewLivestreamHandler(ctrl controllers.LivestreamController) *LivestreamHandler {
	return &LivestreamHandler{
		ctrl: ctrl,
	}
}

func (h *LivestreamHandler) GetLivestreamById(w http.ResponseWriter, r *http.Request) {
	streamId := r.PathValue("livestreamId")
	if len(streamId) == 0 {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("missing livestream id"))
		return
	}

	livestreamUUID, err := uuid.FromString(streamId)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("livestreamId not valid"))
		return
	}

	livestream, err := h.ctrl.GetById(livestreamUUID)
	if err != nil && errors.Is(err, repositories.ErrRecordNotFound) {
		h.WriteErrorResponse(w, http.StatusNotFound, errors.New("livestream not found"))
		return
	} else if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(livestream)
}

func (h *LivestreamHandler) GetLivestreamsOfUser(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("userId")
	if len(userId) == 0 {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("missing user id"))
		return
	}

	userUUID, err := uuid.FromString(userId)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("userId not valid"))
		return
	}

	livestreams, err := h.ctrl.GetByUser(userUUID)
	if err != nil && errors.Is(err, repositories.ErrRecordNotFound) {
		h.WriteErrorResponse(w, http.StatusNotFound, errors.New("user not found"))
		return
	} else if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(livestreams)
}

func (h *LivestreamHandler) CreateLivestream(w http.ResponseWriter, r *http.Request) {
	var body dto.CreateLivestreamRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("error decoding request body: %s", err.Error()))
		return
	}

	if err := utils.Validator.Struct(&body); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("error validating payload: %s", err))
		return
	}

	if body.Title == nil {
		newTitle := "Livestream - " + time.Now().Format(time.RFC3339)
		body.Title = &newTitle
	}

	createdLivestream, err := h.ctrl.Create(body)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createdLivestream)
}

func (h *LivestreamHandler) UpdateLivestream(w http.ResponseWriter, r *http.Request) {
	rawStreamId := r.PathValue("livestreamId")
	streamId, err := uuid.FromString(rawStreamId)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	var requestBody dto.UpdateLivestreamRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("error decoding request body: %s", err.Error()))
		return
	}

	if err := utils.Validator.Struct(&requestBody); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("error validating payload: %s", err))
		return
	}

	if requestBody.Id != streamId {
		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("hey dont try that again: id body and param not equal"))
		return
	}

	updatedLivestream, err := h.ctrl.Update(requestBody)
	if err != nil && errors.Is(err, repositories.ErrRecordNotFound) {
		h.WriteErrorResponse(w, http.StatusNotFound, errors.New("livestream not found"))
		return
	} else if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedLivestream)
}

func (h *LivestreamHandler) DeleteLivestream(w http.ResponseWriter, r *http.Request) {
	rawStreamId := r.PathValue("livestreamId")
	streamId, err := uuid.FromString(rawStreamId)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	err = h.ctrl.Delete(streamId)
	if err != nil && errors.Is(err, repositories.ErrRecordNotFound) {
		h.WriteErrorResponse(w, http.StatusNotFound, errors.New("livestream not found"))
		return
	} else if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
