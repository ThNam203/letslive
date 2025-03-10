package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/lets-live/livestream/dto"
	servererrors "sen1or/lets-live/livestream/errors"
	usergateway "sen1or/lets-live/livestream/gateway/user/http"
	"sen1or/lets-live/livestream/services"
	"strconv"

	"github.com/gofrs/uuid/v5"
)

type LivestreamHandler struct {
	ErrorHandler
	livestreamService services.LivestreamService
	userGateway       usergateway.UserGateway
}

func NewLivestreamHandler(livestreamService services.LivestreamService, userGateway usergateway.UserGateway) *LivestreamHandler {
	return &LivestreamHandler{
		livestreamService: livestreamService,
		userGateway:       userGateway,
	}
}

func (h *LivestreamHandler) GetLivestreamByIdHandler(w http.ResponseWriter, r *http.Request) {
	streamId := r.PathValue("livestreamId")
	if len(streamId) == 0 {
		h.WriteErrorResponse(w, servererrors.ErrInvalidPath)
		return
	}

	livestreamUUID, err := uuid.FromString(streamId)
	if err != nil {
		h.WriteErrorResponse(w, servererrors.ErrInvalidInput)
		return
	}

	livestream, serviceErr := h.livestreamService.GetById(livestreamUUID)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(livestream)
}

func (h *LivestreamHandler) CheckIsUserLivestreamingHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("userId")
	if len(userId) == 0 {
		h.WriteErrorResponse(w, servererrors.ErrInvalidPath)
		return
	}

	userUUID, err := uuid.FromString(userId)
	if err != nil {
		h.WriteErrorResponse(w, servererrors.ErrInvalidInput)
		return
	}

	isLivestreaming, serviceErr := h.livestreamService.CheckIsUserLivestreaming(userUUID)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strconv.FormatBool(isLivestreaming)))
}

func (h *LivestreamHandler) GetLivestreamsOfUserHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("userId")
	if len(userId) == 0 {
		h.WriteErrorResponse(w, servererrors.ErrInvalidPath)
		return
	}

	userUUID, err := uuid.FromString(userId)
	if err != nil {
		h.WriteErrorResponse(w, servererrors.ErrInvalidInput)
		return
	}

	livestreams, serviceErr := h.livestreamService.GetByUser(userUUID)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(livestreams)
}

func (h *LivestreamHandler) GetLivestreamingsHandler(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	var pageNum int
	var err error
	if len(page) == 0 {
		h.WriteErrorResponse(w, servererrors.ErrInvalidInput)
		return
	}

	if pageNum, err = strconv.Atoi(page); err != nil {
		h.WriteErrorResponse(w, servererrors.ErrInvalidInput)
		return
	}

	livestreams, serviceErr := h.livestreamService.GetAllLivestreaming(pageNum)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	res := []dto.GetAllLivestreamingsResponseDTO{}

	for _, live := range livestreams {
		userInfo, err := h.userGateway.GetUserInformationById(context.Background(), live.UserId.String())
		if err == nil {
			res = append(res, dto.GetAllLivestreamingsResponseDTO{
				Id:             live.Id,
				UserId:         live.UserId,
				Username:       userInfo.Username,
				DisplayName:    userInfo.DisplayName,
				ProfilePicture: userInfo.ProfilePicture,
				Title:          live.Title,
				Description:    live.Description,
				ThumbnailURL:   live.ThumbnailURL,
				Status:         live.Status,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *LivestreamHandler) CreateLivestreamHandler(w http.ResponseWriter, r *http.Request) {
	var body dto.CreateLivestreamRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.WriteErrorResponse(w, servererrors.ErrInvalidPayload)
		return
	}

	createdLivestream, err := h.livestreamService.Create(body)
	if err != nil {
		h.WriteErrorResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createdLivestream)
}

func (h *LivestreamHandler) UpdateLivestreamHandler(w http.ResponseWriter, r *http.Request) {
	rawStreamId := r.PathValue("livestreamId")
	streamId, err := uuid.FromString(rawStreamId)
	if err != nil {
		h.WriteErrorResponse(w, servererrors.ErrInvalidInput)
		return
	}
	defer r.Body.Close()

	var requestBody dto.UpdateLivestreamRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		h.WriteErrorResponse(w, servererrors.ErrInvalidPayload)
		return
	}

	updatedLivestream, serviceErr := h.livestreamService.Update(requestBody, streamId)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedLivestream)
}

func (h *LivestreamHandler) DeleteLivestreamHandler(w http.ResponseWriter, r *http.Request) {
	rawStreamId := r.PathValue("livestreamId")
	streamId, err := uuid.FromString(rawStreamId)

	if err != nil {
		h.WriteErrorResponse(w, servererrors.ErrInvalidPath)
		return
	}

	serviceErr := h.livestreamService.Delete(streamId)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
