package handlers

import (
	"encoding/json"
	"net/http"
	"sen1or/letslive/livestream/dto"
	servererrors "sen1or/letslive/livestream/errors"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/services"
	"sen1or/letslive/livestream/types"
	"strconv"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

type LivestreamHandler struct {
	ErrorHandler
	livestreamService services.LivestreamService
}

func NewLivestreamHandler(livestreamService services.LivestreamService) *LivestreamHandler {
	return &LivestreamHandler{
		livestreamService: livestreamService,
	}
}

func (h *LivestreamHandler) CreateLivestreamInternalHandler(w http.ResponseWriter, r *http.Request) {
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

func (h *LivestreamHandler) GetLivestreamByIdPublicHandler(w http.ResponseWriter, r *http.Request) {
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

func (h *LivestreamHandler) GetLivestreamsOfUserPublicHandler(w http.ResponseWriter, r *http.Request) {
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

	livestreams, serviceErr := h.livestreamService.GetByUserPublic(userUUID)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(livestreams)
}

// TODO: paging
func (h *LivestreamHandler) GetAllLivestreamsOfAuthorPrivateHandler(w http.ResponseWriter, r *http.Request) {
	userUUID, err := getUserIdFromCookie(r)
	if err != nil {
		h.WriteErrorResponse(w, servererrors.ErrInvalidInput)
		return
	}

	livestreams, serviceErr := h.livestreamService.GetByUserAuthor(*userUUID)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(livestreams)
}

func (h *LivestreamHandler) GetPopularVODsPublicHandler(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	if len(page) == 0 {
		h.WriteErrorResponse(w, servererrors.ErrInvalidPath)
		return
	}
	pageNum, err := strconv.Atoi(page)
	if err != nil {
		h.WriteErrorResponse(w, servererrors.ErrInvalidInput)
		return
	}

	livestreams, serviceErr := h.livestreamService.GetPopularVODs(pageNum)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(livestreams)
}

func (h *LivestreamHandler) GetLivestreamingsPublicHandler(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(livestreams)
}

func (h *LivestreamHandler) UpdateLivestreamInternalHandler(w http.ResponseWriter, r *http.Request) {
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

	updatedLivestream, serviceErr := h.livestreamService.Update(requestBody, streamId, nil)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedLivestream)
}

func (h *LivestreamHandler) UpdateLivestreamPrivateHandler(w http.ResponseWriter, r *http.Request) {
	userUUID, e := getUserIdFromCookie(r)
	if e != nil {
		h.WriteErrorResponse(w, e)
		return
	}

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

	updatedLivestream, serviceErr := h.livestreamService.Update(requestBody, streamId, userUUID)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedLivestream)
}

func (h *LivestreamHandler) DeleteLivestreamPrivateHandler(w http.ResponseWriter, r *http.Request) {
	rawStreamId := r.PathValue("livestreamId")
	streamId, err := uuid.FromString(rawStreamId)

	if err != nil {
		h.WriteErrorResponse(w, servererrors.ErrInvalidPath)
		return
	}

	userUUID, cErr := getUserIdFromCookie(r)
	if cErr != nil {
		h.WriteErrorResponse(w, cErr)
		return
	}

	serviceErr := h.livestreamService.Delete(streamId, *userUUID)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getUserIdFromCookie(r *http.Request) (*uuid.UUID, *servererrors.ServerError) {
	accessTokenCookie, err := r.Cookie("ACCESS_TOKEN")
	if err != nil || len(accessTokenCookie.Value) == 0 {
		logger.Debugf("missing credentials")
		return nil, servererrors.ErrUnauthorized
	}

	myClaims := types.MyClaims{}

	// the signature should already been checked from the api gateway before going to this
	_, _, err = jwt.NewParser().ParseUnverified(accessTokenCookie.Value, &myClaims)
	if err != nil {
		logger.Debugf("invalid access token: %s", err)
		return nil, servererrors.ErrUnauthorized
	}

	userUUID, err := uuid.FromString(myClaims.UserId)
	if err != nil {
		logger.Debugf("userId not valid")
		return nil, servererrors.ErrUnauthorized
	}

	return &userUUID, nil
}
